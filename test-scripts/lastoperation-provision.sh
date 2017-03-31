#!/bin/bash -e

. shared.sh

curl \
  -H 'X-Broker-API-Version: 2.9' \
  $curlargs \
  $endpoint/broker/my.broker.io/v2/service_instances/$instanceUUID/last_operation'?operation=provisioning'
echo
