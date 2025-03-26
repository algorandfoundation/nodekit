#!/usr/bin/env bash

LOGPFX=$(basename "$0"):
cd "$(dirname "$(realpath "$0")")/.."

source utils/_common.sh

TIMEOUT=$1

if ! ./utils/is_node_running.sh; then
    echo "$LOGPFX Node is not running"
    exit 1
fi

if [ "$TIMEOUT" != "" ]; then
    echo "$LOGPFX Timeout: $TIMEOUT s."
fi

echo -n "$LOGPFX Waiting"

trap 'exit 1' INT

sleep 5

START=$(date +%s)
while ./utils/is_node_syncing.sh; do
    if [[ "$TIMEOUT" != "" ]]; then
        NOW=$(date +%s)
        if [ $(( NOW - START )) -gt $TIMEOUT ]; then
            echo -e "\n$LOGPFX timed out"
            exit 1
        fi
    fi
    echo -n "."
    sleep 2
done

echo ""
echo "$LOGPFX Synced"