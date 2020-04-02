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

docker build -t gcr.io/$PROJECT_ID/samplr-rtr:$UUID -f ./samplr-rtr/Dockerfile ..
docker tag gcr.io/$PROJECT_ID/samplr-rtr:$UUID gcr.io/$PROJECT_ID/samplr-rtr:latest

docker push gcr.io/$PROJECT_ID/samplr-rtr:$UUID
docker push gcr.io/$PROJECT_ID/samplr-rtr:latest

docker build -t gcr.io/$PROJECT_ID/samplr-sprvsr:$UUID -f ./samplr-sprvsr/Dockerfile ..
docker tag gcr.io/$PROJECT_ID/samplr-sprvsr:$UUID gcr.io/$PROJECT_ID/samplr-sprvsr:latest

docker push gcr.io/$PROJECT_ID/samplr-sprvsr:$UUID
docker push gcr.io/$PROJECT_ID/samplr-sprvsr:latest

docker build -t gcr.io/$PROJECT_ID/samplrd:$UUID -f ./samplrd/Dockerfile ..
docker tag gcr.io/$PROJECT_ID/samplrd:$UUID gcr.io/$PROJECT_ID/samplrd:latest

docker push gcr.io/$PROJECT_ID/samplrd:$UUID
docker push gcr.io/$PROJECT_ID/samplrd:latest

(
    cd terraform 
    terraform apply ${auto_approve} -var="image_tag=${UUID}"
)