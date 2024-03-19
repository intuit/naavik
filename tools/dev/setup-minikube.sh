#!/bin/bash

SCRIPT_PATH=`dirname $(readlink -f $0)`
source $SCRIPT_PATH/utils.sh

KUBERNETES_VERSION=v1.29.2
ISTIO_VERSION=1.21.0
CLUSTER_NAME=$1

# Check if minikube is installed
if minikube version | grep -q "minikube version"; then
    echo "Using minikube: $(minikube version)"
else
    echo "Minikube not found. Please install minikube and try again. Run brew install minikube"
    exit 1
fi

if [ -z "$CLUSTER_NAME" ]; then
    echo "Cluster name not provided. Please provide a cluster name as the first argument."
    exit 1
fi

minikube start --memory=8500 --cpus=4 --kubernetes-version=${KUBERNETES_VERSION} -p ${CLUSTER_NAME}  --driver=podman

kubectl config use-context ${CLUSTER_NAME}
kubectl config view --minify > ${CLUSTER_NAME}

if kubectl config current-context | grep -q ${CLUSTER_NAME}; then
    echo "Current context is ${CLUSTER_NAME}"
else 
    echo "Current context must be ${CLUSTER_NAME}. Ensure that you have the correct context selected."; 
    exit 1
fi

# Install istio
# https://istio.io/latest/docs/setup/getting-started/#download
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=${ISTIO_VERSION} TARGET_ARCH=arm64 sh -
./istio-${ISTIO_VERSION}/bin/istioctl install -y --set profile=default
echo "Waiting for istio pods to be ready"
kubectl wait --for=condition=ready pod -l app=istiod --timeout=120s -n istio-system
echo "Istio control plane is ready"

# Install argo-rollouts
# https://argo-rollouts.readthedocs.io/en/stable/installation/
echo "Installing argo-rollouts"
kubectl create namespace argo-rollouts
kubectl apply -n argo-rollouts -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml

printf "${GREEN} Setting up of ${CLUSTER_NAME} cluster is complete.${RESET}"