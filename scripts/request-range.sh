#!/bin/bash

set -eu

FROM=$1
TO=$2
URL='127.0.0.1:8545'

DATA='{
    "jsonrpc": "2.0",
      "method": "statediff_writeStateDiffsInRange",
      "params": ['"$FROM"', '"$TO"', {
          "includeBlock": true,
          "includeReceipts": true,
          "includeTD": true,
          "includeCode": true
        }
      ],
      "id": 1
}'

exec curl -s $URL -X POST -H 'Content-Type: application/json' --data "$DATA"
