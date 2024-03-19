#!/bin/bash

SCRIPT_PATH=`dirname $(readlink -f $0)`
source $SCRIPT_PATH/utils.sh

SECRET_NAMESPACE=admiral

# First cluster is the cluster that will be used for naavik
CLUSTER_NAME=$1

if [ -z "$CLUSTER_NAME" ]; then
    printf "${RED}Cluster name not provided. Please provide a cluster name as the first argument.${END}"
    exit 1
fi

kubectl config use-context ${CLUSTER_NAME}
kubectl config view --minify > ${CLUSTER_NAME}

if kubectl config current-context | grep -q ${CLUSTER_NAME}; then
    printf "${GREEN}Current context is ${CLUSTER_NAME}${END}"
else 
    printf "${RED}Current context must be ${CLUSTER_NAME}. Ensure that you have the correct context selected.${END}"; 
    exit 1
fi

# Install required CRDs
kubectl apply -f ./manifests/crds


# Create secret namespace and secret
for CLUSTER in "$@"
do
    if [ -f $CLUSTER ]; then
        printf "${GREEN}Adding ${CLUSTER} cluster to remote cluster secret namespace.${END}"
    else
        printf "${RED}${CLUSTER} K8s config does not exist in $(pwd)/$CLUSTER. Create the a minikube cluster by running ./tools/dev-setup/setup-minikube.sh $CLUSTER ${END}"
        exit 1
    fi 
    kubectl create namespace ${SECRET_NAMESPACE}
    kubectl create secret generic -n ${SECRET_NAMESPACE} ${CLUSTER} --from-file=${CLUSTER}
    kubectl label secret -n ${SECRET_NAMESPACE} ${CLUSTER} admiral.io/sync=true
done

# Start naavik
go run ./cmd/naavik/main.go --log_level trace --log_color=true --kube_config=$(pwd)/${CLUSTER_NAME} --config_resolver=secret --state_checker=none --config_path ./config/config.yaml --sync_period 60s
