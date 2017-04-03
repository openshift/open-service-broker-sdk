#!/bin/bash

targetNamespace=brokersdk

# Installer creates namespace of its choice
oc new-project ${targetNamespace}

# Installer creates `brokersdk` service
oc create -f resources/service.yaml

# Installer creates `brokersdk` service account
oc create -f resources/sa.yaml


# Installer binds “known” roles to SA user
oc create -f resources/roles.yaml || true

oadm policy add-cluster-role-to-user system:auth-delegator -n ${targetNamespace} -z brokersdk
oc create policybinding kube-system -n kube-system
oadm policy add-role-to-user extension-apiserver-authentication-reader -n kube-system --role-namespace=kube-system system:serviceaccount:${targetNamespace}:brokersdk

# Create the brokersdk replication controller
oc create -f resources/rc.yaml
