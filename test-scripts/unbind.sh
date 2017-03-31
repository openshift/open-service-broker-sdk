#!/bin/bash -e

. shared.sh

curl \
  -X DELETE \
  -H 'X-Broker-API-Version: 2.9' \
  $curlargs \
  $endpoint/broker/my.broker.io/v2/service_instances/$instanceUUID/service_bindings/$bindingUUID
echo
