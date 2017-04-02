#!/bin/bash -e

. shared.sh

curl \
  -H 'X-Broker-API-Version: 2.9' \
  $curlargs \
  $endpoint/broker/sdkbroker.broker.io/v2/catalog
echo
