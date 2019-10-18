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

type ServiceNamer func(repos.TrackedRepository) (string, error)
type DeploymentNamer func(repos.TrackedRepository) (string, error)
type DeploymentBuilder func(repos.TrackedRepository) (*appsv1.Deployment, error)
type ServiceBuilder func(repos.TrackedRepository) (*apiv1.Service, error)
type DeploymentPrep func(repos.TrackedRepository) error

type K8sConfiguration struct {
	ServiceNamer ServiceNamer
	DeploymentNamer DeploymentNamer
	ServiceBuilder ServiceBuilder
	DeploymentBuilder DeploymentBuilder
	PreDeploy DeploymentPrep
}


// Supervisor exists to supervise a set of deployments
type Supervisor interface {
	Supervise(string, func(error)) error
}