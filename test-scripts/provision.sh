#!/bin/bash -e

. shared.sh

req="{
  \"plan_id\": \"$planUUID\",
  \"service_id\": \"$serviceUUID\",
  \"parameters\": {
    \"FOO\": \"BAR\"
  },
  \"accepts_incomplete\": true
}"

curl \
  -X PUT \
  -H 'X-Broker-API-Version: 2.9' \
  -H 'Content-Type: application/json' \
  -d "$req" \
  $curlargs \
  $endpoint/broker/sdkbroker.broker.io/v2/service_instances/$instanceUUID
echo
curl $curlargs $endpoint/apis/sdkbroker.broker.k8s.io/v1alpha1/namespaces/brokersdk/serviceinstances
echo
