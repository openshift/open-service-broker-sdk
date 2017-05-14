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

## Developing a new broker

### Setting up your fork

1) Fork the repository
2) Clone your fork
3) Run `hack/fork-rename.sh <your github org name> <your github repo name>`
    *  this will rename all the package imports to match your project

### Customizing the repo

In writing your own Broker implementation, you will primarily be concerned with the following 3 areas:

* Broker API implementation - pkg/openservicebroker/operations
* Broker State Resources - pkg/apis/broker, pkg/registry/broker/serviceinstance
* Provision Controller implementation - pkg/controller


#### Broker API Implementation

The main logic of what your broker will do when `provision`/`deprovision`/`bind`/`unbind`/`catalog`/`lastoperation` requests
are made is implemented in the `pkg/openservicebroker/operations` package.  These functions
are automatically bound to the appropriate API endpoints by the Broker SDK and will be invoked
by the `Service Catalog`.

You can customize this logic to implement whatever actions you want to take in response to requests from the Service Catalog.

The provided implementation executes the following flow:

1) On provision, a `ServiceInstance` resource is create.  This is a Kubernetes resource that is provided by the Broker SDK
itself and stored in a local etcd instance that will be running in the same pod as the Broker.  A controller(see next section) will
observe the new ServiceInstance and process it to complete the provision operation, and add the Ready condition

2) On a lastoperation request, the `ServiceInstance` object conditions will be checked to see if it is Pending, Ready, or Failed.

3) On bind, the broker confirms the presence of a `ServiceInstance` with a matching uuid, and if found, returns some credential
information associated with the ServiceInstance.

4) On unbind, the broker confirms the presence of a `ServiceInstance` with a matching uuid and based the success/failure response
on the existence of that object.

5) On deprovision, the broker deletes the `ServiceInstance` object.  The controller could, but does not currently, take some
additional action upon receiving the deletion event.

6) On a catalog request, the broker returns a valid, but hardcoded, list of catalog entries.

#### Broker State Resources

The Broker SDK defines its own API group resource which is served by the API server which is part of the Broker process.
This API group defines one resource type, a `ServiceInstance`.  This resource is used to store the state of provision requests.
Defining additional resources or customization to this resource are expected during usage of the SDK.  Since the Broker calls
itself to access these resources, no special permissions should need to be granted to users to access the resources, they
can be considered internal only.  The resources are backed by etcd storage in a local etcd process running in a second container
with in the Broker SDK pod that is defined.  Real world implementations will need to take into account the backup/restore
strategies for this data.


#### Controller Implementation

The Broker SDK provides a single controller which watches for `ServiceInstance` objects.  This allows it to implement an
asynchronous provision flow in which the provision call returns immediately after defining a new `ServiceInstance` (
effectively a service provision request).  The controller is free to implement any business logic necessary to handle
the request, such as creating backend resources needed by the service instance being provisioned.  Once the service
instance is prepared, the controller updates the `ServiceInstance` object condition to indicate it is ready.  This
information is used by the `lastoperation` API when checking on the state of a provision request.

On provision, the current controller implementation simply updates the `ServiceInstance` to indicate it is ready.
On deletion, the controller simply logs the event but a real Broker would be expected to cleanup the provisioned resources
here.
