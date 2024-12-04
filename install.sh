#!/usr/bin/env bash

set -euo pipefail

os=$(uname -ms)
# TODO: replace with algorandfoundation org and publicly host script
release="https://github.com/awesome-algorand/hack-tui/releases/download/"
version="v1.0.0-beta.1"

if [[ ${OS:-} = Windows_NT ]]; then
  echo "Unsupported platform"
  exit 1
fi


case $os in
'Darwin x86_64')
    target=algorun-amd64-darwin
    ;;
'Darwin arm64')
    target=algorun-arm64-darwin
    ;;
'Linux aarch64' | 'Linux arm64')
    target=algorun-arm64-linux
    ;;
'Linux x86_64' | *)
    target=algorun-amd64-linux
    ;;
esac

echo "Downloading: $release/$version/$target"
curl --fail --location --progress-bar --output algorun "$release/$version/$target"
