#!/usr/bin/env bash

LOGPFX=$(basename "$0"):
cd "$(dirname "$(realpath "$0")")/.."

set -e

source utils/_common.sh

addr=$1
amt=$2

echo -n "$LOGPFX Waiting for ${addr:0:6}.. to reach $amt microalgo"
bal=0
while [ "$(./utils/get_balance.sh "$addr")" -lt "$amt" ]; do
    echo -n "."
    sleep 30;
done
