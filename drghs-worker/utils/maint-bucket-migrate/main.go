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
	"log"
	"os"
	"path/filepath"
	"strings"

	"devrel/cloud/devrel-github-service/repos"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	settingsBucket = flag.String("settings-bucket", "devrel-prod-settings", "Google Cloud Storage bucket to use for settings storage")
	reposFileName  = flag.String("file", "backed_repos.json", "The list of public repos")
	fromPrefix     = flag.String("from-prefix", "mtr-b-", "The prefix to the bucket to copy from")
	toPrefix       = flag.String("to-prefix", "mtr-p-", "The prefix of the bucket to copy to")
	projectID      = flag.String("gcp-project", "", "The GCP Project this is using")
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	if *projectID == "" {
		log.Fatalf("--gcp-project is required")
	}

	if *settingsBucket == "" {
		log.Fatalf("--settings-bucket is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repoList := repos.NewBucketRepo(*settingsBucket, *reposFileName)
	_, err = repoList.UpdateTrackedRepos(ctx)
	if err != nil {
		log.Fatal(err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, tr := range repoList.GetTrackedRepos() {
		log.Printf("Processing repo: %v:", tr.String())
		fromBucketName := bucketName(tr, *fromPrefix)
		toBucketName := bucketName(tr, *toPrefix)

		fromBucket := client.Bucket(fromBucketName)
		toBucket := client.Bucket(toBucketName)

		dn, err := deploymentName(tr)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Deleting deployment: %v", dn)
		err = clientset.AppsV1().Deployments("default").Delete(dn, &metav1.DeleteOptions{})
		if err != nil && errors.IsNotFound(err) {
			err = nil
		}
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Emptying Bucket: %v", toBucketName)
		oi := toBucket.Objects(ctx, nil)
		for {
			objAttrs, err := oi.Next()
			if err == iterator.Done {
				break
			}
			if err == storage.ErrBucketNotExist {
				toBucket.Create(ctx, *projectID, nil)
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Deleting object: %v", objAttrs.Name)
			o := toBucket.Object(objAttrs.Name)
			err = o.Delete(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Printf("Copying from %v", fromBucketName)
		oi = fromBucket.Objects(ctx, nil)
		for {
			objAttrs, err := oi.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Copying object: %v", objAttrs.Name)
			fO := fromBucket.Object(objAttrs.Name)
			tO := toBucket.Object(objAttrs.Name)
			tO.CopierFrom(fO).Run(ctx)

		}
	}

	log.Print("Deleting the supervisor pod (to restart the deleted pods")
	err = clientset.CoreV1().Pods("default").Delete("maintnerd-sprvsr", &metav1.DeleteOptions{})
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Deleted!")

	log.Print("Finished!")
}

func bucketName(t repos.TrackedRepository, prefix string) string {
	bld := strings.Builder{}
	bld.WriteString(prefix)
	s := t.RepoSha()
	bld.WriteString(s)
	return bld.String()
}

func deploymentName(t repos.TrackedRepository) (string, error) {
	return strings.ToLower(fmt.Sprintf("mtr-d-%s", t.RepoSha())), nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
