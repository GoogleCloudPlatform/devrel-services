// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sprvsr

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/GoogleCloudPlatform/devrel-services/repos"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	labelgenkeyfmt = "%v-sprvsr-autogen"
	labelgenvalue  = "true"
)

// Ensure k8supervisor is Supervisor
var _ Supervisor = &k8supervisor{}

type k8supervisor struct {
	mu        sync.RWMutex
	log       *logrus.Logger
	clientset kubernetes.Interface

	// CRUD Operations on K8s Objects

	serviceNamer      ServiceNamer
	deploymentNamer   DeploymentNamer
	serviceBuilder    ServiceBuilder
	deploymentPrep    DeploymentPrep
	deploymentBuilder DeploymentBuilder

	// The list of repositories to track
	repoList repos.RepoList

	// A unique name (per k8s cluster) for your application to supervise
	labelgenkey string
}

// ServiceNamer is called to determine what to name a Service Given a TrackedRepository
type ServiceNamer func(repos.TrackedRepository) (string, error)

// DeploymentNamer is called to determine what to name a Deployment Given a TrackedRepository
type DeploymentNamer func(repos.TrackedRepository) (string, error)

// DeploymentBuilder builds a Deployment based on the given TrackedRepository
type DeploymentBuilder func(repos.TrackedRepository) (*appsv1.Deployment, error)

// ServiceBuilder buidls a Service based on the given TrackedRepository
type ServiceBuilder func(repos.TrackedRepository) (*apiv1.Service, error)

// DeploymentPrep is called before building a deployment. This can be
// used to provision additional resources before the Deployment is applied
type DeploymentPrep func(repos.TrackedRepository) error

// K8sConfiguration is a struct to describe the set of operations
// a K8SSupervisor needs to manage a cluster
type K8sConfiguration struct {
	ServiceNamer      ServiceNamer
	DeploymentNamer   DeploymentNamer
	ServiceBuilder    ServiceBuilder
	DeploymentBuilder DeploymentBuilder
	PreDeploy         DeploymentPrep
}

// NewK8sSupervisor creates a new supervisor backed by Kubernetes
func NewK8sSupervisor(log *logrus.Logger, clientset kubernetes.Interface, kconfig K8sConfiguration,
	rl repos.RepoList,
	appid string) (Supervisor, error) {
	return newK8sSupervisor(log, clientset, kconfig, rl, appid)
}

func newK8sSupervisor(log *logrus.Logger, clientset kubernetes.Interface, kconfig K8sConfiguration,
	rl repos.RepoList,
	appid string) (*k8supervisor, error) {

	lblkey := fmt.Sprintf(labelgenkeyfmt, appid)

	return &k8supervisor{
		clientset:         clientset,
		log:               log,
		serviceNamer:      kconfig.ServiceNamer,
		deploymentNamer:   kconfig.DeploymentNamer,
		serviceBuilder:    kconfig.ServiceBuilder,
		deploymentBuilder: kconfig.DeploymentBuilder,
		deploymentPrep:    kconfig.PreDeploy,
		repoList:          rl,
		labelgenkey:       lblkey,
	}, nil
}

// Supervise registers an http server on the given address
// and error handler. This watches the Kubernetes cluster for
// changes and enforces them with the /update route
func (s *k8supervisor) Supervise(address string, handle func(error)) error {
	go s.updateCorpusRepoList(context.Background(), handle)

	s.log.Debugf("creating router")
	// Send everything through Mux
	r := mux.NewRouter()

	s.log.Debugf("handling update")
	r.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		s.updateCorpusRepoList(r.Context(), handle)
	}).Methods("GET", "POST")

	s.log.Debugf("handling healthz")
	r.HandleFunc("/_healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	// Add middleware support
	n := negroni.New()
	l := negroni.NewLogger()
	n.Use(l)
	n.Use(negroni.NewRecovery())
	n.UseHandler(r)

	return http.ListenAndServe(address, n)
}

func (s *k8supervisor) updateCorpusRepoList(ctx context.Context, handle func(error)) {
	s.mu.RLock()
	changed, err := s.repoList.UpdateTrackedRepos(ctx)
	if err != nil {
		handle(err)
		s.mu.RUnlock()
		return
	}
	s.mu.RUnlock()
	if !changed {
		s.log.Debug("skipping updating corpus repo list. unchanged")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	trackedRepos := s.repoList.GetTrackedRepos()

	s.log.Debugf("got n tracked repos: %v", len(trackedRepos))

	trSet := make(map[repos.TrackedRepository]bool)
	for _, tr := range trackedRepos {
		trSet[tr] = true
	}

	s.log.Debugf("trSet: %v", trSet)

	deploymentsSet := make(map[repos.TrackedRepository]bool)
	servicesSet := make(map[repos.TrackedRepository]bool)

	// Store this as a variable here in the event we want this configurable
	ns := apiv1.NamespaceDefault
	labelSelector := fmt.Sprintf("%v=%v", s.labelgenkey, labelgenvalue)

	deployments, err := s.clientset.AppsV1().Deployments(ns).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		handle(err)
		return
	}
	for _, deployment := range deployments.Items {
		// Now we have the deployment.... inspect the
		// labels in the deployment to get the owner and repository
		if _, ok := deployment.Labels["owner"]; !ok {
			// Log here?
			continue
		}
		if _, ok := deployment.Labels["repository"]; !ok {
			// Log here?
			continue
		}

		o := deployment.Labels["owner"]
		r := deployment.Labels["repository"]

		deploymentsSet[repos.TrackedRepository{
			Owner: o,
			Name:  r,
		}] = true
	}

	s.log.Debugf("have deployments from k8s: %v", deploymentsSet)

	services, err := s.clientset.CoreV1().Services(ns).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		handle(err)
		return
	}
	for _, service := range services.Items {
		// Now we have the service... inspect the
		// labels on the service to get the owner and repository
		if _, ok := service.Labels["owner"]; !ok {
			continue
		}
		if _, ok := service.Labels["repository"]; !ok {
			continue
		}

		o := service.Labels["owner"]
		r := service.Labels["repository"]
		servicesSet[repos.TrackedRepository{
			Owner: o,
			Name:  r,
		}] = true
	}

	s.log.Debugf("have services from k8s: %v", servicesSet)

	// Delete Services before deployments
	servicesToDelete := setDifference(servicesSet, trSet)
	s.log.Debugf("have services to delete: %v", servicesToDelete)
	for td := range servicesToDelete {
		// Delete the service
		sn, err := s.serviceNamer(td)
		if err != nil {
			handle(err)
			continue
		}

		err = s.clientset.CoreV1().Services(ns).Delete(sn, &metav1.DeleteOptions{})
		if err != nil {
			handle(err)
		}
	}

	deploymentsToDelete := setDifference(deploymentsSet, trSet)
	s.log.Debugf("have deployments to delete: %v", deploymentsToDelete)
	for td := range deploymentsToDelete {
		// Delete the deployment
		dn, err := s.deploymentNamer(td)
		if err != nil {
			handle(err)
			continue
		}
		err = s.clientset.AppsV1().Deployments(ns).Delete(dn, &metav1.DeleteOptions{})
		if err != nil {
			handle(err)
		}
	}

	// Add deployments before services
	deploymentsToAdd := setDifference(trSet, deploymentsSet)
	s.log.Debugf("have deployments to add: %v", deploymentsToAdd)
	for ta := range deploymentsToAdd {
		// Add the deployment
		if err := s.deploymentPrep(ta); err != nil {
			handle(err)
			continue
		}

		if err := createDeployment(s.clientset, ns, s.labelgenkey, s.deploymentBuilder, s.deploymentNamer, ta); err != nil {
			handle(err)
		}
	}

	servicesToAdd := setDifference(trSet, servicesSet)
	s.log.Debugf("have services to add: %v", servicesToAdd)
	for ta := range servicesToAdd {
		// Add the service
		if err := createService(s.clientset, ns, s.labelgenkey, s.serviceBuilder, ta); err != nil {
			handle(err)
		}
		s.log.Debugf("created service for %v", ta)
	}
}

func createDeployment(cs kubernetes.Interface, ns string, lblkey string, db DeploymentBuilder, dn DeploymentNamer, ta repos.TrackedRepository) error {
	d, err := db(ta)
	if err != nil {
		return err
	}
	dname, err := dn(ta)
	if err != nil {
		return err
	}

	// Give the deployment our uinque label
	if d.ObjectMeta.Labels == nil {
		d.ObjectMeta.Labels = make(map[string]string, 0)
	}
	d.ObjectMeta.Labels[lblkey] = labelgenvalue

	// Add Owner and repository labels to the deployment
	d.ObjectMeta.Labels["owner"] = ta.Owner
	d.ObjectMeta.Labels["repository"] = ta.Name

	// Give the pods our uinque label
	if d.Spec.Template.ObjectMeta.Labels == nil {
		d.Spec.Template.ObjectMeta.Labels = make(map[string]string, 0)
	}
	d.Spec.Template.ObjectMeta.Labels[lblkey] = labelgenvalue

	// Add Owner and repository labels to the pods
	d.Spec.Template.ObjectMeta.Labels["owner"] = ta.Owner
	d.Spec.Template.ObjectMeta.Labels["repository"] = ta.Name

	_, err = cs.AppsV1().Deployments(ns).Create(d)

	if err != nil && err.Error() == fmt.Sprintf("deployments.apps: \"%v\" already exists", dname) {
		err = nil
	}
	return err
}

func createService(cs kubernetes.Interface, ns string, labelgenkey string, sb ServiceBuilder, ta repos.TrackedRepository) error {
	svc, err := sb(ta)
	if err != nil {
		return err
	}

	if svc.ObjectMeta.Labels == nil {
		svc.ObjectMeta.Labels = make(map[string]string, 0)
	}
	// Give the object our unique identifier
	svc.ObjectMeta.Labels[labelgenkey] = labelgenvalue

	svc.ObjectMeta.Labels["owner"] = ta.Owner
	svc.ObjectMeta.Labels["repository"] = ta.Name

	_, err = cs.CoreV1().Services(ns).Create(svc)
	return err
}

// Returns the difference between this set
// and other. The returned set will contain
// all elements of this set that are not also
// elements of other.
func setDifference(this, other map[repos.TrackedRepository]bool) map[repos.TrackedRepository]bool {
	diff := make(map[repos.TrackedRepository]bool)
	for k := range this {
		if _, ok := other[k]; !ok {
			diff[k] = true
		}
	}
	return diff
}
