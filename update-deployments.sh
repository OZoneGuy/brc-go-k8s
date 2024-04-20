#!/bin/bash

DIRS="file-reader parser collector"

for dir in $DIRS; do
  cd $dir
  kubectl delete -f deployment.yaml --wait=true
  sleep 1
  minikube image rm $dir
  minikube image build -t $dir .
  kubectl apply -f deployment.yaml --wait=true
  cd ..
done
