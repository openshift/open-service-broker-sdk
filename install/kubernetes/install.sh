#!/bin/sh -e

targetNamespace=brokersdk

# Installer creates namespace of its choice
kubectl create ns ${targetNamespace}

# Installer creates `brokersdk` service
kubectl create -f ../service.yaml -n ${targetNamespace}

# Installer creates `brokersdk` service account
kubectl create -f ../sa.yaml -n ${targetNamespace}


# Installer binds “known” roles to SA user
kubectl create -f ./roles.yaml

# Create the brokersdk replication controller
kubectl create -f ../rc.yaml -n ${targetNamespace}
