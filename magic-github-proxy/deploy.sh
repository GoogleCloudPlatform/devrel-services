#!/bin/bash
set -euxo pipefail

echo Active project is $(gcloud config list --format 'value(core.project)')

auto_approve=""

if [ "$#" -eq  "0" ] || [[ "$1" -ne "-y"  ]]; then    
    echo Do you wish to continue? [yN]
    read ans 
    if [[ $ans != "y" ]]; then
        exit
    fi
elif [[ "$1" -eq "-y" ]]; then
    auto_approve="--auto-approve"
fi 

PROJECT_ID=$(gcloud config list --format 'value(core.project)')

UUID=$(git rev-parse HEAD)

docker build -t gcr.io/$PROJECT_ID/magic-github-proxy:$UUID -f Dockerfile .
docker tag gcr.io/$PROJECT_ID/magic-github-proxy:$UUID gcr.io/$PROJECT_ID/magic-github-proxy:latest

docker push gcr.io/$PROJECT_ID/magic-github-proxy:$UUID
docker push gcr.io/$PROJECT_ID/magic-github-proxy:latest

(
    cd terraform 
    terraform init && \
    terraform apply ${auto_approve} -var="image_tag=${UUID}"
)