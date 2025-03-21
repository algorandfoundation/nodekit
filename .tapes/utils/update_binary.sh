#!/usr/bin/env bash

set -e

cd "$(dirname "$(realpath "$0")")/.."

LOGPFX=$(basename "$0"):

source utils/_common.sh

# ensure have sudo access
sudo echo "" > /dev/null

mkdir -p tmp

SHASUMS_URL="https://fnet.algorand.green/static/SHA256SUMS"
DOWNLOAD_URL_PREFIX="https://fnet.algorand.green/static"

TMPDIR="bin"

echo "$LOGPFX downloading SHA256SUMS"

UPDATED=0
UPDATED_NAMES=()

TMPFILE_SHA="$TMPDIR/SHA256SUMS"
curl -sf "$SHASUMS_URL" -o "$TMPFILE_SHA"

for file in algod goal; do
    LATEST_SHA=$(grep $file "$TMPFILE_SHA" | cut -d\  -f1)
    echo "$LOGPFX Latest $file: $LATEST_SHA"

    BINARY_PATH="bin/$file"

    CURRENT_SHA=$(sha256 "$BINARY_PATH")
    echo "$LOGPFX Installed $file: $CURRENT_SHA"

    if [[ "$CURRENT_SHA" = "$LATEST_SHA" ]]; then
        echo "$LOGPFX $file: no update"
    else
        TMPFILE_BINARY="$TMPDIR/$file"
        echo "$LOGPFX $file: update available"

        echo "$LOGPFX Downloading $file"
        curl -sf "$DOWNLOAD_URL_PREFIX/$file" -o "$TMPFILE_BINARY"

        DOWNLOADED_SHA=$(sha256 "$TMPFILE_BINARY")
        if [[ "$DOWNLOADED_SHA" = "$LATEST_SHA" ]]; then
            echo "$LOGPFX SHA256 hash verified."
            UPDATED_NAMES+=("$file")
            UPDATED=$(( UPDATED + 1 ))
        else
            echo -e "$LOGPFX ERROR: Incorrect $file hash.\nExpected: $LATEST_SHA\nFound: $DOWNLOADED_SHA"
            echo "$LOGPFX Files left in $TMPDIR"
            trap - EXIT
            echo "$LOGPFX ERROR"
            exit 1
        fi
    fi
done