# Open Service Broker SDK

A skeleton project for creating new service brokers that implement the
[Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker).
Our goal is eventually to make this project a full-featured SDK that handles
the subtleties of implementing the API and allows broker authors to focus on
the business logic of provisioning and binding to their services.

## Purpose

This is intended as a starting point for new broker implementations that will
run inside a kubernetes cluster.  The intent is for broker implementers to
fork this repository and fill in their own broker specific logic/resource
definitions into the skeleton that is provided.

We are specifically most interested in brokers that will integrate with the
[Kubernetes Service Catalog](https://github.com/kubernetes-incubator/service-catalog).

## Current Status

This project is currently in the pre-alpha phase of its existence.  The
current usage pattern for this project is to fork/clone it and modify it in
order to implement your own broker.

Our next step is to take existing brokers and reimplement them using this
project. This will allow us to find pitfalls and bugs as well as determining
what the lifecycle of a project based on this SDK should look like.

## Running the example

```
$ make images
$ # start up an openshift/kube cluster
$ # have admin credentials
$ cd test-scripts
$ ./install-broker.sh
$ ./provision.sh
$ ./bind.sh
$ ./unbind.sh
$ ./deprovision.sh
```
