#!/bin/bash

SCRIPT_PATH=`dirname $(readlink -f $0)`
source $SCRIPT_PATH/utils.sh

# Install required CRDs
printf "${GREEN}Installing CRDs${END}"
kubectl apply -f ./manifests/crds