#!/usr/bin/env bash

set -e

if [ "$DEBUG" = "1" ]; then
  set -x
fi

if [ "$ALGORAND_DATA" != "/algod/data" ]; then
  echo "Do not override 'ALGORAND_DATA' environment variable."
  exit 1
fi

/node/run/start_empty.sh &
/node/run/start_fast_catchup.sh &
/node/run/start_dev.sh