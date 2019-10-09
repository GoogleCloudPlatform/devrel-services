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
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"devrel/cloud/devrel-github-service/repos"
	"devrel/cloud/devrel-github-service/sprvsr"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/storage"
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
	listen           = flag.String("listen", ":6343", "listen address")
	verbose          = flag.Bool("verbose", false, "enable verbose debug output")
	settingsBucket   = flag.String("settings-bucket", "cdpe-maintner-settings", "Google Cloud Storage bucket to use for settings storage")
	reposFileName    = flag.String("repos-file", "", "File that contains the list of repositories")
	projectID        = flag.String("gcp-project", "", "The GCP Project this is using")
	githubSecretName = flag.String("github-secret", "", "The name of the secret that contains the GitHub tokens")
	sasecretname     = flag.String("service-account-secret", "", "The name of the ServiceAccount for our Pods to run as")
	mimagename       = flag.String("maint-image-name", "", "The name of the image to run maintner")
	mutationBucket   = flag.String("mutation-bucket", "", "The bucket to store mutation data")
)

// Config
var (
	repoList    repos.RepoList
	errorClient *errorreporting.Client
	config      *rest.Config
	mu          sync.RWMutex
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

	log.Out = os.Stdout
}

func main() {
	// Set log to Stdout. Default for log is Stderr
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

	if *githubSecretName == "" {
		err := fmt.Errorf("must provide --github-secret")
		logAndPrintError(err)
		log.Fatal(err)
	}

	if *sasecretname == "" {
		err := fmt.Errorf("must provide --service-account-secret")
		logAndPrintError(err)
		log.Fatal(err)
	}

	if *mimagename == "" {
		err := fmt.Errorf("must provide --maint-image-name")
		logAndPrintError(err)
		log.Fatal(err)
	}
	if *mutationBucket == "" {
		err := fmt.Errorf("must provide --mutation-bucket")
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

	preDeploy := func(ta repos.TrackedRepository) error {
		return nil
	}

	cdeployment := func(ta repos.TrackedRepository) (*appsv1.Deployment, error) {
		githubsecretkey, err := getGithubSecretName(cs, apiv1.NamespaceDefault)
		if err != nil {
			return nil, err
		}
		return buildDeployment(*sasecretname, *githubSecretName, githubsecretkey, ta)
	}

	kcfg := sprvsr.K8sConfiguration{
		ServiceNamer:      serviceName,
		DeploymentNamer:   deploymentName,
		ServiceBuilder:    buildService,
		DeploymentBuilder: cdeployment,
		PreDeploy:         preDeploy,
	}

	super, err := sprvsr.NewK8sSupervisor(log, cs, kcfg, repoList, "maintner")
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

func getGithubSecretName(cs *kubernetes.Clientset, ns string) (string, error) {
	// We need some information to add our deployments... in particular, we
	// need the set of github keys we have available as secrets
	availablesecrets, err := getTokenNames(cs, ns, *githubSecretName)
	if err != nil {
		logAndPrintError(err)
		return "", err
	}
	if len(availablesecrets) < 1 {
		err := fmt.Errorf("no secrets stored in %v", *githubSecretName)
		logAndPrintError(err)
		return "", err
	}
	log.Debugf("have secrets to vend: %v", len(availablesecrets))

	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)

	// Get a random key for our secrets
	// TODO(colnnelson): if a Tracked Repository specifies
	// a particular key to use, look that up and use it instead
	//
	// rng.Intn() returns a random int32 between 0 and n, so we need
	// to guard against 0
	idx := 0
	if len(availablesecrets) != 1 {
		idx = rng.Intn(len(availablesecrets) - 1)
	}
	keyName := availablesecrets[idx]
	return keyName, nil
}

func serviceName(t repos.TrackedRepository) (string, error) {
	return strings.ToLower(fmt.Sprintf("mtr-s-%s", t.RepoSha())), nil
}

func deploymentName(t repos.TrackedRepository) (string, error) {
	return strings.ToLower(fmt.Sprintf("mtr-d-%s", t.RepoSha())), nil
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
					TargetPort: intstr.FromInt(80),
				},
			},
			Selector: map[string]string{
				"app": dep,
			},
			Type: "ClusterIP",
		},
	}, nil
}

func buildDeployment(sasecretname, githubsecretname, githubsecretkey string, ta repos.TrackedRepository) (*appsv1.Deployment, error) {
	dep, err := deploymentName(ta)
	if err != nil {
		return nil, err
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
						"app": dep,
					},
				},
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: "gcp-sa",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: sasecretname,
								},
							},
						},
					},
					Containers: []apiv1.Container{
						apiv1.Container{
							Name:            "maintnerd",
							Image:           *mimagename,
							ImagePullPolicy: "Always",
							Command: []string{
								"/cmd",
								fmt.Sprintf("--bucket=%v", bucketName(ta)),
								"--verbose",
								"--token=$(GITHUB_TOKEN)",
								"--listen=:80",
								fmt.Sprintf("--gcp-project=%v", *projectID),
								fmt.Sprintf("--owner=%v", ta.Owner),
								fmt.Sprintf("--repo=%v", ta.Name),
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							LivenessProbe: &apiv1.Probe{
								Handler: apiv1.Handler{
									HTTPGet: &apiv1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(80),
									}},
								InitialDelaySeconds: 10,
								PeriodSeconds:       3,
							},
							Env: []apiv1.EnvVar{
								apiv1.EnvVar{
									Name:  "GOOGLE_APPLICATION_CREDENTIALS",
									Value: "/var/secrets/google/key.json",
								},
								apiv1.EnvVar{
									Name: "GITHUB_TOKEN",
									ValueFrom: &apiv1.EnvVarSource{
										SecretKeyRef: &apiv1.SecretKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: githubsecretname,
											},
											Key: githubsecretkey,
										},
									},
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      "gcp-sa",
									MountPath: "/var/secrets/google",
								},
							},
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									// Our application does not need "that" much CPU.
									// For context, if unspecified, GKE applies a default request of "100m"
									apiv1.ResourceCPU:    resource.MustParse("50m"),
									apiv1.ResourceMemory: resource.MustParse("160M"),
								},
								Limits: apiv1.ResourceList{
									// Limit the CPU ask
									apiv1.ResourceCPU: resource.MustParse("1000m"),
									// As of this writing the "monolithic" maintner service is
									// consuming 3 GB of RAM, and peaked at 3.4 GB.
									apiv1.ResourceMemory: resource.MustParse("2G"),
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func getTokenNames(clientset *kubernetes.Clientset, ns, secretname string) ([]string, error) {
	secret, err := clientset.CoreV1().Secrets(ns).Get(secretname, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	avail := make([]string, len(secret.Data))
	idx := 0
	for k := range secret.Data {
		avail[idx] = k
		idx++
	}
	return avail, nil
}

func createBucket(ctx context.Context, ta repos.TrackedRepository, projectID string) error {
	sc, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	name := bucketName(ta)
	b := sc.Bucket(name)
	err = b.Create(ctx, projectID, nil)
	if err != nil && err.Error() == "googleapi: Error 409: You already own this bucket. Please select another name., conflict" {
		err = nil
	}
	return err
}

func bucketName(t repos.TrackedRepository) string {
	return path.Join(*mutationBucket)
}

func int32Ptr(i int32) *int32 { return &i }
