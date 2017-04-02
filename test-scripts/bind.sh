#!/bin/bash -e

. shared.sh

req="{
  \"plan_id\": \"$planUUID\",
  \"service_id\": \"$serviceUUID\",
  \"parameters\": {
    \"template.openshift.io/requester-username\": \"$requesterUsername\"
  }
}"

curl \
  -X PUT \
  -H 'X-Broker-API-Version: 2.9' \
  -H 'Content-Type: application/json' \
  -d "$req" \
  $curlargs \
  $endpoint/broker/sdkbroker.broker.io/v2/service_instances/$instanceUUID/service_bindings/$bindingUUID
echo
