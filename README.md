# BrokerSDK

A skeleton project for creating new service brokers that follow the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker)

## Purpose

This is intended as a starting point for new broker implementations that will run inside a kubernetes cluster.  The intent is for
broker implementers to fork this repository and fill in their own broker specific logic/resource definitions into the 
skeleton that is provided.

## Running the example

```
$ make images
$ start up an openshift/kube cluster
$ have admin credentials
$ cd test-scripts
$ ./install-broker.sh
$ ./provision.sh
$ ./bind.sh
$ ./unbind.sh
$ ./deprovision.sh
```
