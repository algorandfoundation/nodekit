#!/usr/bin/env bash

LOGPFX=$(basename "$0"):
cd "$(dirname "$(realpath "$0")")/.."

set -e

bootstrap=_algobootstrap._tcp.fnet.algorand.green
port=8184

resps=$(dig +short srv $bootstrap | cut -d\  -f4 | sed 's/\.$//' | shuf)

echo "$LOGPFX Resolved $(echo -e "$resps" | wc -l) relays" >&2

for relayHostname in $resps; do
    if genesis=$(curl -s "http://$relayHostname:$port/genesis"); then
        echo "$LOGPFX Got genesis, $(echo -e "$genesis" | wc -l) lines" >&2
        echo -e "$genesis"
        exit 0
    fi
done

echo "$LOGPFX Failed to get genesis"

exit 1