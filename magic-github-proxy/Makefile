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
# The KMS Settings
KMS_LOCATION ?= global
KMS_KEY_RING ?= magic-github-proxy
KMS_KEY ?= enc-at-rest
# The GCS Bucket Settings
GCS_BUCKET_NAME ?= magic-github-proxy
# Asymetric key files
PRIVATE_KEY_FILE ?= private.pem.enc
CERTIFICATE_FILE ?= public.x509.cer.enc

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
	docker build -t magic-github-proxy:dev .

.PHONY: deploy
deploy: check-project check-deploy ## Deploys the app to the cluster
	gcloud builds submit \
    --config=cloudbuild.yaml \
    --substitutions _ZONE=$(GCP_CLUSTER_ZONE),_CLUSTER=$(GCP_CLUSTER_NAME),_KMS_LOCATION=$(KMS_LOCATION),_KMS_KEY_RING=$(KMS_KEY_RING),_KMS_KEY=$(KMS_KEY),_BUCKET_NAME=$(GCS_BUCKET_NAME),_PRIVATE_FILE=$(PRIVATE_KEY_FILE),_CERTIFICATE_FILE=$(CERTIFICATE_FILE) \
	.

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

.PHONY: help
help: ## Prints help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
