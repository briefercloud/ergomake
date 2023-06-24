#!/bin/bash

set -xe

minikube delete
minikube start --insecure-registry=host.minikube.internal:5001 --nodes 4 --network-plugin=cni --cni=calico --kubernetes-version="v1.24.3" --memory=max --cpus=max
minikube addons enable ingress

kubectl label nodes minikube "preview.ergomake.dev/role=core"
kubectl label nodes minikube-m02 "preview.ergomake.dev/role=preview"
kubectl label nodes minikube-m03 "preview.ergomake.dev/role=system"
kubectl label nodes minikube-m04 "preview.ergomake.dev/role=build"

kubectl label nodes minikube-m02 "preview.ergomake.dev/logs=true"
kubectl label nodes minikube-m04 "preview.ergomake.dev/logs=true"

kubectl taint nodes minikube "preview.ergomake.dev/domain=core:NoSchedule"
kubectl taint nodes minikube-m02 "preview.ergomake.dev/domain=previews:NoSchedule"
kubectl taint nodes minikube-m04 "preview.ergomake.dev/domain=build:NoSchedule"


echo "Install the ergomake helm chart to create the build namespace and service account."
