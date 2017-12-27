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

sleep 10;

kubectl -n kube-system annotate deployments kube-dns hpk-num-nodes='1'
kubectl -n kube-system annotate deployments kube-dns hpk-base-image='ubuntu'
kubectl -n kube-system annotate deployments kube-dns hpk-pull-policy='Always'
kubectl -n kube-system annotate deployments kube-dns hpk-default-namespace='kube-public'
kubectl -n kube-system annotate deployments kube-dns hpk-max-cpu='200m'
kubectl -n kube-system annotate deployments kube-dns hpk-max-memory='128Mi'
#pwd=$(pwd)

#minikube ssh "cd ${pwd} && ./build.sh"
