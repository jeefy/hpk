#!/bin/bash

CLUSTER=$(minikube status)
STATUS=$?
if [ "$STATUS" -lt 1 ]; then
  echo "Existing minikube detected.";
  minikube delete
else
  echo "Starting fresh!";
fi

minikube start --memory 4096

docker rm -f kjob-mongo
docker run --name kjob-mongo -p 27017:27017 -d mongo

sleep 12;

source hack/bootstrap-config.sh
#pwd=$(pwd)

#minikube ssh "cd ${pwd} && ./build.sh"
