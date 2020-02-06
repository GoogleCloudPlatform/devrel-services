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
	"github.com/GoogleCloudPlatform/devrel-services/repos"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

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

// Supervisor exists to supervise a set of deployments
type Supervisor interface {
	Supervise(string, func(error)) error
}
