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

docker build -t gcr.io/$PROJECT_ID/maintner-rtr:$UUID -f ./maintner-rtr/Dockerfile ..
docker tag gcr.io/$PROJECT_ID/maintner-rtr:$UUID gcr.io/$PROJECT_ID/maintner-rtr:latest

docker push gcr.io/$PROJECT_ID/maintner-rtr:$UUID
docker push gcr.io/$PROJECT_ID/maintner-rtr:latest

docker build -t gcr.io/$PROJECT_ID/maintner-sprvsr:$UUID -f ./maintner-sprvsr/Dockerfile ..
docker tag gcr.io/$PROJECT_ID/maintner-sprvsr:$UUID gcr.io/$PROJECT_ID/maintner-sprvsr:latest

docker push gcr.io/$PROJECT_ID/maintner-sprvsr:$UUID
docker push gcr.io/$PROJECT_ID/maintner-sprvsr:latest

docker build -t gcr.io/$PROJECT_ID/maintnerd:$UUID -f ./maintnerd/Dockerfile ..
docker tag gcr.io/$PROJECT_ID/maintnerd:$UUID gcr.io/$PROJECT_ID/maintnerd:latest

docker push gcr.io/$PROJECT_ID/maintnerd:$UUID
docker push gcr.io/$PROJECT_ID/maintnerd:latest

(
    cd terraform 
    terraform apply ${auto_approve} -var="image_tag=${UUID}"
)