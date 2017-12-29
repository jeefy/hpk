#!/bin/bash

kubectl -n kube-system annotate deployments kube-dns hpk-num-nodes='1'
kubectl -n kube-system annotate deployments kube-dns hpk-base-image='ubuntu'
kubectl -n kube-system annotate deployments kube-dns hpk-pull-policy='Always'
kubectl -n kube-system annotate deployments kube-dns hpk-default-namespace='kube-public'
kubectl -n kube-system annotate deployments kube-dns hpk-max-cpu='200m'
kubectl -n kube-system annotate deployments kube-dns hpk-max-memory='128Mi'
kubectl -n kube-system annotate deployments kube-dns hpk-cost-cpu='0.0000023' # Cost of 1% core / s
kubectl -n kube-system annotate deployments kube-dns hpk-cost-memory='0.000000000023' # Cost of 1MB RAM / s
