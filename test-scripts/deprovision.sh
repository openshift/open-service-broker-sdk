#!/bin/bash -e

. shared.sh

curl \
  -X DELETE \
  -H 'X-Broker-API-Version: 2.9' \
  $curlargs \
  $endpoint/broker/my.broker.io/v2/service_instances/$instanceUUID'?accepts_incomplete=true'
echo
curl $curlargs $endpoint/apis/generic.broker.k8s.io/v1alpha1/namespaces/brokersdk/serviceinstances
echo
