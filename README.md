# BrokerSDK

A skeleton project for creating new service brokers that follow the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker)

## Purpose

This is intended as a starting point for new broker implementations that will run inside a kubernetes cluster.

## Getting started

$ make images
$ start up an openshift/kube cluster
$ have admin credentials
$ cd test-scripts
$ ./install-broker.sh
$ ./provision.sh
$ ./bind.sh
$ ./unbind.sh
$ ./deprovision.sh
