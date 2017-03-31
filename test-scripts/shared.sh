#!/bin/bash -e

# id of the service that is being requested from the broker
serviceUUID=d5d242c1-164e-11e7-8714-0242ac110001
# id of the plan that is being requested from the broker
planUUID=d5d242c1-164e-11e7-8714-0242ac110003

# the id for the service instance we are provisioning/acting on 
# (initially provided by the service catalog on provision)
instanceUUID=d5d242c1-164e-11e7-8714-0242ac110002

# id for the binding we are acting on (initially provided
# by the service catalog on bind)
bindingUUID=d5d242c1-164e-11e7-8714-0242ac110004

# kube service where our broker is running
broker_service_ip=`oc get svc brokersdk -o jsonpath={.spec.clusterIP}`
endpoint=https://${broker_service_ip}:8443

# pass -k to curl because we don't have a valid TLS cert
curlargs=${curlargs--k}
