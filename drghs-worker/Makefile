# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


#  __     __         _       _     _
#  \ \   / /_ _ _ __(_) __ _| |__ | | ___ ___
#   \ \ / / _` | '__| |/ _` | '_ \| |/ _ \ __|
#    \ V / (_| | |  | | (_| | |_) | |  __\__ \
#     \_/ \__,_|_|  |_|\__,_|_.__/|_|\___|___/
#

#
# All of the following can be overwritten with environemt variables
# or passed through directly when invoking the relevent Make targets
#

# The (gcloud) test cluster that is being worked against
GCP_CLUSTER_NAME ?= devrel-services
GCP_CLUSTER_ZONE ?= us-central1-a
# The service account to run as
SERVICE_ACCOUNT_SECRET_NAME ?= service-account-maintnerd
# Bucket settings for Repositories
GCS_BUCKET_NAME ?= devrel-dev-settings
REPOS_FILE_NAME ?= public_repos.json
# The name of the GKE Secret that contains GitHub Access Tokens
GITHUB_SECRET_NAME ?= github-token
# The prefix to use for Buckets
MUTATION_BUCKET ?= mtr-dev


#   _____                    _
#  |_   _|_ _ _ __ __ _  ___| |_ ___
#    | |/ _` | '__/ _` |/ _ \ __/ __|
#    | | (_| | | | (_| |  __/ |_\__ \
#    |_|\__,_|_|  \__, |\___|\__|___/
#                 |___/
#

.PHONY: all
all: help

.PHONY: build
build: ## Builds the docker image locally
	docker build -t maintnerd:dev -f maintnerd/Dockerfile ../
	docker build -t maintnerd-rtr:dev -f maintner-rtr/Dockerfile ../
	docker build -t maintnerd-sprvsr:dev -f maintner-sprvsr/Dockerfile ../

.PHONY: test
test: ## Runs the unit tests locally
	go test ./test

.PHONY: deploy
deploy: check-project check-deploy ## Deploys the app to the GKE cluster
	gcloud builds submit \
	--config=cloudbuild.yaml \
	--timeout=1h \
	--substitutions _ZONE=$(GCP_CLUSTER_ZONE),_CLUSTER=$(GCP_CLUSTER_NAME),_BUCKET_NAME=$(GCS_BUCKET_NAME),_REPO_FILE_NAME=$(REPOS_FILE_NAME),_SERVICE_ACCOUNT_SECRET=$(SERVICE_ACCOUNT_SECRET_NAME),_GITHUB_SECRET_NAME=$(GITHUB_SECRET_NAME),_PREFIX=$(MUTATION_BUCKET) \
	../

.PHONY: check-deploy
check-deploy: printvars
	 @echo -n "Are you sure? [yN] " && read ans && [ $$ans == y ]

.PHONY: check-project
check-project:
	@echo "Active project is: $$(gcloud config list --format 'value(core.project)')"
	@echo -n "Are you sure? [yN] " && read ans && [ $$ans == y ]

.PHONY: printvars
printvars:
	@$(foreach V,$(sort $(.VARIABLES)),\
		$(if $(filter-out environment% default automatic,\
		$(origin $V)),$(info $V=$($V) ($(value $V)))))

.PHONY: update-deps
update-deps: ## Updates the go dependencies in the Dockerfile
	go install golang.org/x/build/cmd/gitlock
	gitlock --ignore=nothing --update=Dockerfile ./cmd

.PHONY: help
help: ## Prints help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
