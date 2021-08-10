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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/GoogleCloudPlatform/devrel-services/repos"
	"github.com/GoogleCloudPlatform/devrel-services/sprvsr"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/profiler"
	"github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Flags
var (
	listen         = flag.String("listen", ":6343", "listen address")
	verbose        = flag.Bool("verbose", false, "enable verbose debug output")
	settingsBucket = flag.String("settings-bucket", "cdpe-maintner-settings", "Google Cloud Storage bucket to use for settings storage")
	reposFileName  = flag.String("repos-file", "", "File that contains the list of repositories")
	projectID      = flag.String("gcp-project", "", "The GCP Project this is using")
	simagename     = flag.String("samplr-image-name", "", "The name of the image to run samplr")
	sasecretname   = flag.String("service-account-secret", "", "The name of the ServiceAccount for our Pods to run as")
)

// Config
var (
	repoList    repos.RepoList
	errorClient *errorreporting.Client
	config      *rest.Config
	mu          sync.RWMutex
)

// Const
const (
	samplrbackendport = 8080
)

// Log
var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}

	// Set log to Stdout. Default for log is Stderr
	log.Out = os.Stdout
}

func main() {
	flag.Parse()

	if *verbose == true {
		log.Level = logrus.TraceLevel
	}

	ctx := context.Background()

	if *projectID == "" {
		log.Fatal("must provide --gcp-project")
	}

	var err error
	errorClient, err = errorreporting.NewClient(ctx, *projectID, errorreporting.Config{
		ServiceName: "devrel-github-services",
		OnError: func(err error) {
			log.Errorf("Could not report error: %v", err)
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer errorClient.Close()

	if err := profiler.Start(profiler.Config{
		Service:        "samplr-sprvsr",
		ServiceVersion: "0.0.1",
		MutexProfiling: true,
	}); err != nil {
		logAndPrintError(fmt.Errorf("error initializing profiler: %v", err))
	}

	if *settingsBucket == "" {
		err := fmt.Errorf("must provide --settings-bucket")
		logAndPrintError(err)
		log.Fatal(err)
	}

	if *reposFileName == "" {
		err := fmt.Errorf("must provide --repos-file")
		logAndPrintError(err)
		log.Fatal(err)
	}

	if *sasecretname == "" {
		err := fmt.Errorf("must provide --service-account-secret")
		logAndPrintError(err)
		log.Fatal(err)
	}

	if *simagename == "" {
		err := fmt.Errorf("must provide --maint-image-name")
		logAndPrintError(err)
		log.Fatal(err)
	}

	// Init k8s info
	// creates the in-cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		logAndPrintError(err)
		log.Fatal(err)
	}

	// We need to interface with the k8s api
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		logAndPrintError(err)
		log.Fatal(err)
	}

	repoList = repos.NewBucketRepo(*settingsBucket, *reposFileName)

	bd := func(ta repos.TrackedRepository) (*appsv1.Deployment, error) {
		return buildDeployment(ta)
	}

	kcfg := sprvsr.K8sConfiguration{
		ServiceNamer:      serviceName,
		DeploymentNamer:   deploymentName,
		ServiceBuilder:    buildService,
		DeploymentBuilder: bd,
		PreDeploy:         preDeploy,
		ShouldDeploy:      shouldDeploy,
	}

	super, err := sprvsr.NewK8sSupervisor(log, cs, kcfg, repoList, "samplr")
	if err != nil {
		logAndPrintError(err)
		log.Fatal(err)
	}

	log.Fatal(super.Supervise(*listen, logAndPrintError))
}

func logAndPrintError(err error) {
	errorClient.Report(errorreporting.Entry{
		Error: err,
	})
	log.Error(err)
}

func serviceName(t repos.TrackedRepository) (string, error) {
	return strings.ToLower(fmt.Sprintf("smp-s-%s", t.RepoSha())), nil
}

func deploymentName(t repos.TrackedRepository) (string, error) {
	return strings.ToLower(fmt.Sprintf("smp-d-%s", t.RepoSha())), nil
}

func buildService(ta repos.TrackedRepository) (*apiv1.Service, error) {
	svc, err := serviceName(ta)
	if err != nil {
		return nil, err
	}
	dep, err := deploymentName(ta)
	if err != nil {
		return nil, err
	}

	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   svc,
			Labels: map[string]string{},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				apiv1.ServicePort{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(samplrbackendport),
				},
			},
			Selector: map[string]string{
				"app": dep,
			},
			Type: "ClusterIP",
		},
	}, nil
}

func buildDeployment(ta repos.TrackedRepository) (*appsv1.Deployment, error) {
	dep, err := deploymentName(ta)
	if err != nil {
		return nil, err
	}
	defaultBranch := "master"
	if ta.DefaultBranch != "" {
		defaultBranch = ta.DefaultBranch
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: dep,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": dep,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":    dep,
						"branch": defaultBranch,
					},
				},
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{},
					Containers: []apiv1.Container{
						apiv1.Container{
							Name:            "samplrd",
							Image:           *simagename,
							ImagePullPolicy: "Always",
							Command: []string{
								"/samplrd",
								fmt.Sprintf("--listen=:%v", samplrbackendport),
								fmt.Sprintf("--owner=%v", ta.Owner),
								fmt.Sprintf("--repo=%v", ta.Name),
								fmt.Sprintf("--branch=%v", defaultBranch),
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: samplrbackendport,
								},
							},
							Env:          []apiv1.EnvVar{},
							VolumeMounts: []apiv1.VolumeMount{},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									// Our application does not need "that" much CPU.
									// For context, if unspecified, k8s applies a default request of "100m"
									apiv1.ResourceCPU:    resource.MustParse("50m"),
									apiv1.ResourceMemory: resource.MustParse("160M"),
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: resource.MustParse("2.5G"),
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func preDeploy(ta repos.TrackedRepository) error {
	return nil
}

func shouldDeploy(ta repos.TrackedRepository) bool {
	return ta.IsTrackingSamples
}

func int32Ptr(i int32) *int32 { return &i }
