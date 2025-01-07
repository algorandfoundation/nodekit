#!/usr/bin/env bash

set -euo pipefail



BANNER='   _____  .__                __________
   /  _  \ |  |    ____   ____\______   \__ __  ____
  /  /_\  \|  |   / ___\ /  _ \|       _/  |  \/    \
 /    |    \  |__/ /_/  >  <_> )    |   \  |  /   |  \
 \____|__  /____/\___  / \____/|____|_  /____/|___|  /
         \/     /_____/               \/           \/ '
  
os=$(uname -ms)
release="https://github.com/algorandfoundation/algorun-tui/releases/download"
version="v1.0.0-beta.3"

Red=''
Green=''
Yellow=''
Blue=''
Opaque=''
Bold_White=''
Bold_Green=''
Reset=''

if [[ -t 1 ]]; then
    Reset='\033[0m'
    Red='\033[0;31m'
    Green='\033[0;32m'
    Yellow='\033[0;33m'
    Blue='\033[0;34m'
    Bold_Green='\033[1;32m'
    Bold_White='\033[1m'
    Opaque='\033[0;2m'
    echo -e "${Blue} ${BANNER} ${Reset}"
fi

success() {
  echo -e "${Green}$@ ${Reset}"
}

info() {
  echo -e "${Opaque}$@ ${Reset}"
}

warn() {
  echo -e "${Yellow}WARN${Reset}: ${Opaque}$@ ${Reset}"
}

error() {
    echo -e "${Red}ERROR${Reset}:" "${Yellow}" "$@" "${Reset}" >&2
    exit 1
}

if [ -f nodekit ]; then
    error "An nodekit file already exists in the current directory. Delete or rename it before installing."
fi


if [[ ${OS:-} = Windows_NT ]]; then
  error "Unsupported platform"
fi

trap "warn SIGINT received." int
trap "info Exiting the installation" exit

case $os in
'Darwin x86_64')
    target=nodekit-amd64-darwin
    ;;
'Darwin arm64')
    target=nodekit-arm64-darwin
    ;;
'Linux aarch64' | 'Linux arm64')
    target=nodekit-arm64-linux
    ;;
'Linux x86_64' | *)
    target=nodekit-amd64-linux
    ;;
esac
 
echo -e "${Opaque}Downloading:${Reset}${Bold_White} $target $version${Reset}"
curl --fail --location --progress-bar --output nodekit "$release/$version/$target" ||
  error "Failed to download ${target} from ${release}"

chmod +x nodekit

trap - int
trap - exit

success "Downloaded: ${Bold_Green}algorun ${version} 🎉${Reset}"
info "Explore nodekit by starting here:"
echo "./nodekit --help"
echo ""
info "Starting nodekit bootstrap"
echo "./algorun bootstrap"

./algorun bootstrap
