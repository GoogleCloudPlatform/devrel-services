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
	"testing"

	"github.com/GoogleCloudPlatform/devrel-services/repos"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSetDifference(t *testing.T) {
	cases := []struct {
		Name  string
		Left  map[string]repos.TrackedRepository
		Right map[string]repos.TrackedRepository
		Want  map[repos.TrackedRepository]bool
	}{
		{
			Name:  "Empty left empty right yields empty",
			Left:  map[string]repos.TrackedRepository{},
			Right: map[string]repos.TrackedRepository{},
			Want:  map[repos.TrackedRepository]bool{},
		},
		{
			Name:  "Empty left full right yields empty",
			Left:  map[string]repos.TrackedRepository{},
			Right: map[string]repos.TrackedRepository{"foo/bar": repos.TrackedRepository{Owner: "foo", Name: "bar"}},
			Want:  map[repos.TrackedRepository]bool{},
		},
		{
			Name:  "Full left empty right yields Left",
			Left:  map[string]repos.TrackedRepository{"foo/bar": repos.TrackedRepository{Owner: "foo", Name: "bar"}},
			Right: map[string]repos.TrackedRepository{},
			Want:  map[repos.TrackedRepository]bool{repos.TrackedRepository{Owner: "foo", Name: "bar"}: true},
		},
		{
			Name:  "Full left full right yields Left",
			Left:  map[string]repos.TrackedRepository{"foo/bar": repos.TrackedRepository{Owner: "foo", Name: "bar"}},
			Right: map[string]repos.TrackedRepository{"baz/biz": repos.TrackedRepository{Owner: "baz", Name: "biz"}},
			Want:  map[repos.TrackedRepository]bool{repos.TrackedRepository{Owner: "foo", Name: "bar"}: true},
		},
		{
			Name:  "Equal Left and Right yields empty",
			Left:  map[string]repos.TrackedRepository{"foo/bar": repos.TrackedRepository{Owner: "foo", Name: "bar"}},
			Right: map[string]repos.TrackedRepository{"foo/bar": repos.TrackedRepository{Owner: "foo", Name: "bar"}},
			Want:  map[repos.TrackedRepository]bool{},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := setDifference(c.Left, c.Right)
			if diff := cmp.Diff(c.Want, got); diff != "" {
				t.Errorf("Test: %v Repositories differ (-want +got)\n%s", c.Name, diff)
			}
		})
	}
}

type fakeRepoList struct {
	RepoStack [][]repos.TrackedRepository
	idx       int
}

func (f *fakeRepoList) UpdateTrackedRepos(context.Context) (bool, error) {
	if f.idx < len(f.RepoStack)-1 {
		f.idx++
	}
	return true, nil
}
func (f *fakeRepoList) GetTrackedRepos() []repos.TrackedRepository {
	return f.RepoStack[f.idx]
}

const ns = apiv1.NamespaceDefault

func TestUpdateAddsDeploymentsAndServices(t *testing.T) {
	log := logrus.New()
	clientSet := fake.NewSimpleClientset()
	config := K8sConfiguration{
		ServiceNamer:    func(repos.TrackedRepository) (string, error) { return "foo", nil },
		DeploymentNamer: func(repos.TrackedRepository) (string, error) { return "foo", nil },
		ServiceBuilder: func(repos.TrackedRepository) (*apiv1.Service, error) {
			return &apiv1.Service{}, nil
		},
		DeploymentBuilder: func(repos.TrackedRepository) (*appsv1.Deployment, error) {
			return &appsv1.Deployment{}, nil
		},
		PreDeploy:    func(repos.TrackedRepository) error { return nil },
		ShouldDeploy: func(tr repos.TrackedRepository) bool { return tr.IsTrackingSamples },
	}
	repoList := &fakeRepoList{
		idx: -1,
		RepoStack: [][]repos.TrackedRepository{
			{
				repos.TrackedRepository{
					Owner:             "beep",
					Name:              "boop",
					IsTrackingSamples: false,
				},
				repos.TrackedRepository{
					Owner:             "foo",
					Name:              "bar",
					IsTrackingSamples: true,
				},
				repos.TrackedRepository{
					Owner:             "beep",
					Name:              "blarp",
					IsTrackingSamples: false,
				},
			},
		},
	}
	appid := "testapp"

	spr, err := newK8sSupervisor(log, clientSet, config, repoList, appid)
	if err != nil {
		t.Errorf("Got an error making a new supervisor: %v", err)
	}

	ctx := context.Background()
	spr.updateCorpusRepoList(ctx, func(error) {})

	services, err := clientSet.CoreV1().Services(ns).List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("Got an error listing services: %v", err)
	}

	if len(services.Items) != 1 {
		t.Errorf("Wanted %v services. Got %v", 1, len(services.Items))
	}

	deployments, err := clientSet.AppsV1().Deployments(ns).List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("Got an error listing deployments: %v", err)
	}

	if len(deployments.Items) != 1 {
		t.Errorf("Wanted %v deployments. Got: %v", 1, len(deployments.Items))
	}
}

func TestUpdateMultipleLeavesThingsAlone(t *testing.T) {
	log := logrus.New()
	clientSet := fake.NewSimpleClientset()
	config := K8sConfiguration{
		ServiceNamer:    func(repos.TrackedRepository) (string, error) { return "foo", nil },
		DeploymentNamer: func(repos.TrackedRepository) (string, error) { return "foo", nil },
		ServiceBuilder: func(repos.TrackedRepository) (*apiv1.Service, error) {
			return &apiv1.Service{}, nil
		},
		DeploymentBuilder: func(repos.TrackedRepository) (*appsv1.Deployment, error) {
			return &appsv1.Deployment{}, nil
		},
		PreDeploy:    func(repos.TrackedRepository) error { return nil },
		ShouldDeploy: func(tr repos.TrackedRepository) bool { return tr.IsTrackingSamples },
	}
	repoList := &fakeRepoList{
		idx: -1,
		RepoStack: [][]repos.TrackedRepository{
			{
				repos.TrackedRepository{
					Owner:             "foo",
					Name:              "bar",
					IsTrackingSamples: true,
					DefaultBranch:     "master",
				},
				// Should be skipped
				repos.TrackedRepository{
					Owner:             "beep",
					Name:              "boop",
					IsTrackingSamples: false,
				},
			},
		},
	}
	appid := "testapp"

	spr, err := newK8sSupervisor(log, clientSet, config, repoList, appid)
	if err != nil {
		t.Errorf("Got an error making a new supervisor: %v", err)
	}

	ctx := context.Background()

	spr.updateCorpusRepoList(ctx, func(error) {})
	spr.updateCorpusRepoList(ctx, func(error) {})
	spr.updateCorpusRepoList(ctx, func(error) {})

	acts := clientSet.Actions()

	ncreate := 0
	for _, a := range acts {
		if a.GetVerb() == "delete" {
			t.Errorf("Did not expect to delete a resource: %v", a)
		} else if a.GetVerb() == "create" {
			ncreate++
		}
	}
	if ncreate != 2 {
		t.Errorf("Wanted %v Created. Got %v", 2, ncreate)
	}
}

func TestNewReposAreAdded(t *testing.T) {
	log := logrus.New()
	clientSet := fake.NewSimpleClientset()
	config := K8sConfiguration{
		ServiceNamer: func(a repos.TrackedRepository) (string, error) {
			return fmt.Sprintf("s-%v", a.RepoSha()), nil
		},
		DeploymentNamer: func(a repos.TrackedRepository) (string, error) {
			return fmt.Sprintf("d-%v", a.RepoSha()), nil
		},
		ServiceBuilder: func(repos.TrackedRepository) (*apiv1.Service, error) {
			return &apiv1.Service{}, nil
		},
		DeploymentBuilder: func(repos.TrackedRepository) (*appsv1.Deployment, error) {
			return &appsv1.Deployment{}, nil
		},
		PreDeploy:    func(repos.TrackedRepository) error { return nil },
		ShouldDeploy: func(tr repos.TrackedRepository) bool { return tr.IsTrackingSamples },
	}
	repoList := &fakeRepoList{
		idx: -1,
		RepoStack: [][]repos.TrackedRepository{
			{
				repos.TrackedRepository{
					Owner:             "foo",
					Name:              "bar",
					IsTrackingSamples: true,
					DefaultBranch:     "main",
				},
				repos.TrackedRepository{
					Owner:             "beep",
					Name:              "boop",
					IsTrackingSamples: false,
					DefaultBranch:     "main",
				},
			},
			{
				repos.TrackedRepository{
					Owner:             "foo",
					Name:              "bar",
					IsTrackingSamples: true,
					DefaultBranch:     "main",
				},
				repos.TrackedRepository{
					Owner:             "beep",
					Name:              "boop",
					IsTrackingSamples: false,
					DefaultBranch:     "main",
				},
				repos.TrackedRepository{
					Owner:             "baz",
					Name:              "biz",
					IsTrackingSamples: true,
					DefaultBranch:     "main",
				},
			},
		},
	}
	appid := "testapp"

	spr, err := newK8sSupervisor(log, clientSet, config, repoList, appid)
	if err != nil {
		t.Errorf("Got an error making a new supervisor: %v", err)
	}

	ctx := context.Background()
	spr.updateCorpusRepoList(ctx, func(error) {})

	ncreate := 0
	ndelete := 0
	for _, a := range clientSet.Actions() {
		if a.GetVerb() == "create" {
			ncreate++
		} else if a.GetVerb() == "delete" {
			ndelete++
		} else {
			t.Logf("Got verb %v", a.GetVerb())
		}
	}
	// Want 2 as we are creating one service and one deployment
	if ncreate != 2 {
		t.Errorf("Wanted %v Created. Got %v", 2, ncreate)
	}
	if ndelete != 0 {
		t.Errorf("Wanted %v Deleted. Got %v", 0, ndelete)
	}

	spr.updateCorpusRepoList(ctx, func(error) {})

	ncreate = 0
	ndelete = 0
	for _, a := range clientSet.Actions() {
		if a.GetVerb() == "create" {
			ncreate++
		} else if a.GetVerb() == "delete" {
			ndelete++
		} else {
			t.Logf("Got verb %v", a.GetVerb())
		}
	}
	// Want 4 as we are created 1 more service and 1 more deployment.
	if ncreate != 4 {
		t.Errorf("Wanted %v Created. Got %v", 4, ncreate)
	}
	if ndelete != 0 {
		t.Errorf("Wanted %v Deleted. Got %v", 0, ndelete)
	}
}

func TestDeletedReposAreRemoved(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	clientSet := fake.NewSimpleClientset()
	config := K8sConfiguration{
		ServiceNamer: func(a repos.TrackedRepository) (string, error) {
			return fmt.Sprintf("s-%v", a.RepoSha()), nil
		},
		DeploymentNamer: func(a repos.TrackedRepository) (string, error) {
			return fmt.Sprintf("d-%v", a.RepoSha()), nil
		},
		ServiceBuilder: func(a repos.TrackedRepository) (*apiv1.Service, error) {
			return &apiv1.Service{ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("s-%v", a.RepoSha()),
			}}, nil
		},
		DeploymentBuilder: func(a repos.TrackedRepository) (*appsv1.Deployment, error) {
			return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("d-%v", a.RepoSha()),
			}}, nil
		},
		PreDeploy:    func(repos.TrackedRepository) error { return nil },
		ShouldDeploy: func(tr repos.TrackedRepository) bool { return tr.IsTrackingSamples },
	}
	repoList := &fakeRepoList{
		idx: -1,
		RepoStack: [][]repos.TrackedRepository{
			{
				repos.TrackedRepository{
					Owner:             "foo",
					Name:              "bar",
					IsTrackingSamples: true,
					DefaultBranch:     "main",
				},
				repos.TrackedRepository{
					Owner:             "baz",
					Name:              "biz",
					IsTrackingSamples: true,
					DefaultBranch:     "main",
				},
				repos.TrackedRepository{
					Owner:             "beep",
					Name:              "boop",
					IsTrackingSamples: false,
					DefaultBranch:     "main",
				},
			},
			{
				repos.TrackedRepository{
					Owner:             "foo",
					Name:              "bar",
					IsTrackingSamples: true,
					DefaultBranch:     "main",
				},
				repos.TrackedRepository{
					Owner:             "beep",
					Name:              "boop",
					IsTrackingSamples: false,
					DefaultBranch:     "main",
				},
			},
		},
	}
	appid := "testapp"

	spr, err := newK8sSupervisor(log, clientSet, config, repoList, appid)
	if err != nil {
		t.Errorf("Got an error making a new supervisor: %v", err)
	}

	ctx := context.Background()
	spr.updateCorpusRepoList(ctx, func(error) {})

	ncreate := 0
	ndelete := 0
	for _, a := range clientSet.Actions() {
		if a.GetVerb() == "create" {
			ncreate++
		} else if a.GetVerb() == "delete" {
			ndelete++
		}
		t.Logf("Got action: %v %v", a.GetResource(), a.GetVerb())
	}
	// Want 4 as we are creating 2 services and 2 deployments
	if ncreate != 4 {
		t.Errorf("Wanted %v Created. Got %v", 4, ncreate)
	}
	if ndelete != 0 {
		t.Errorf("Wanted %v Deleted. Got %v", 0, ndelete)
	}

	spr.updateCorpusRepoList(ctx, func(error) {})

	ncreate = 0
	ndelete = 0
	for _, a := range clientSet.Actions() {
		if a.GetVerb() == "create" {
			ncreate++
		} else if a.GetVerb() == "delete" {
			ndelete++
		}
		t.Logf("Got action: %v %v", a.GetResource(), a.GetVerb())

	}
	// Want 2 deletes for the service and Deployment
	if ndelete != 2 {
		t.Errorf("Wanted %v Deleted. Got %v", 2, ndelete)
	}
}

func TestUpdatedImagesAreReplaced(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	appid := "testapp"

	// This desployment has an image that is DIFFERENT than the
	// desired on to be built, so it should be deleted and replaced.
	existingDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "d-foo-bar",
			Labels: map[string]string{
				"owner":                  "foo",
				"repository":             "bar",
				"testapp-sprvsr-autogen": "true",
			},
			Namespace: apiv1.NamespaceDefault,
		},
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{},
					Containers: []apiv1.Container{
						apiv1.Container{
							Image: "foo-bar:baz",
						},
					},
				},
			},
		},
	}

	clientSet := fake.NewSimpleClientset(existingDeployment)
	config := K8sConfiguration{
		ServiceNamer: func(a repos.TrackedRepository) (string, error) {
			return fmt.Sprintf("s-%v-%v", a.Owner, a.Name), nil
		},
		DeploymentNamer: func(a repos.TrackedRepository) (string, error) {
			return fmt.Sprintf("d-%v-%v", a.Owner, a.Name), nil
		},
		ServiceBuilder: func(repos.TrackedRepository) (*apiv1.Service, error) {
			return &apiv1.Service{}, nil
		},
		DeploymentBuilder: func(repos.TrackedRepository) (*appsv1.Deployment, error) {
			return &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: apiv1.PodTemplateSpec{
						Spec: apiv1.PodSpec{
							Volumes: []apiv1.Volume{},
							Containers: []apiv1.Container{
								apiv1.Container{
									Image: "foo-bar:biz",
								},
							},
						},
					},
				},
			}, nil
		},
		PreDeploy:    func(repos.TrackedRepository) error { return nil },
		ShouldDeploy: func(tr repos.TrackedRepository) bool { return tr.IsTrackingSamples },
	}
	repoList := &fakeRepoList{
		idx: -1,
		RepoStack: [][]repos.TrackedRepository{
			{
				repos.TrackedRepository{
					Owner:             "foo",
					Name:              "bar",
					IsTrackingSamples: true,
				},
				repos.TrackedRepository{
					Owner:             "beep",
					Name:              "boop",
					IsTrackingSamples: false,
				},
			},
		},
	}

	spr, err := newK8sSupervisor(log, clientSet, config, repoList, appid)
	if err != nil {
		t.Errorf("Got an error making a new supervisor: %v", err)
	}

	ctx := context.Background()
	spr.updateCorpusRepoList(ctx, func(error) {})

	ncreate := 0
	ndelete := 0
	for _, a := range clientSet.Actions() {
		t.Logf("Got verb %v", a.GetVerb())

		if a.GetVerb() == "create" {
			ncreate++
		} else if a.GetVerb() == "delete" {
			ndelete++
		}
	}
	// Want 2 as we are creating one service and one deployment
	if ncreate != 2 {
		t.Errorf("Wanted %v Created. Got %v", 2, ncreate)
	}
	// We should delete the existing deployment
	if ndelete != 1 {
		t.Errorf("Wanted %v Deleted. Got %v", 1, ndelete)
	}
}

func TestChangedBranchesAreReplaced(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	appid := "testapp"

	// This desployment has a branch that is DIFFERENT than the
	// desired on to be built, so it should be deleted and replaced.
	existingDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "d-foo-bar",
			Labels: map[string]string{
				"owner":                  "foo",
				"repository":             "bar",
				"testapp-sprvsr-autogen": "true",
				"branch":                 "branch0",
			},
			Namespace: apiv1.NamespaceDefault,
		},
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{},
					Containers: []apiv1.Container{
						apiv1.Container{
							Image: "foo-bar:baz",
						},
					},
				},
			},
		},
	}

	clientSet := fake.NewSimpleClientset(existingDeployment)
	config := K8sConfiguration{
		ServiceNamer: func(a repos.TrackedRepository) (string, error) {
			return fmt.Sprintf("s-%v-%v", a.Owner, a.Name), nil
		},
		DeploymentNamer: func(a repos.TrackedRepository) (string, error) {
			return fmt.Sprintf("d-%v-%v", a.Owner, a.Name), nil
		},
		ServiceBuilder: func(repos.TrackedRepository) (*apiv1.Service, error) {
			return &apiv1.Service{}, nil
		},
		DeploymentBuilder: func(repos.TrackedRepository) (*appsv1.Deployment, error) {
			return &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "d-foo-bar",
					Labels: map[string]string{
						"branch": "branch1",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: apiv1.PodTemplateSpec{
						Spec: apiv1.PodSpec{
							Volumes: []apiv1.Volume{},
							Containers: []apiv1.Container{
								apiv1.Container{
									Image: "foo-bar:baz",
								},
							},
						},
					},
				},
			}, nil
		},
		PreDeploy:    func(repos.TrackedRepository) error { return nil },
		ShouldDeploy: func(tr repos.TrackedRepository) bool { return tr.IsTrackingSamples },
	}
	repoList := &fakeRepoList{
		idx: -1,
		RepoStack: [][]repos.TrackedRepository{
			{
				repos.TrackedRepository{
					Owner:             "foo",
					Name:              "bar",
					IsTrackingSamples: true,
					DefaultBranch:     "branch1",
				},
				repos.TrackedRepository{
					Owner:             "beep",
					Name:              "boop",
					IsTrackingSamples: false,
				},
			},
		},
	}

	spr, err := newK8sSupervisor(log, clientSet, config, repoList, appid)
	if err != nil {
		t.Errorf("Got an error making a new supervisor: %v", err)
	}

	ctx := context.Background()
	spr.updateCorpusRepoList(ctx, func(error) {})

	ncreate := 0
	ndelete := 0
	for _, a := range clientSet.Actions() {
		t.Logf("Got verb %v", a.GetVerb())

		if a.GetVerb() == "create" {
			ncreate++
		} else if a.GetVerb() == "delete" {
			ndelete++
		}
	}
	// Want 2 as we are creating one service and one deployment
	if ncreate != 2 {
		t.Errorf("Wanted %v Created. Got %v", 2, ncreate)
	}
	// We should delete the existing deployment
	if ndelete != 1 {
		t.Errorf("Wanted %v Deleted. Got %v", 1, ndelete)
	}
}
