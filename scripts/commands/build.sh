#!/usr/bin/env bash

# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"


echoTitle "Building code"
set -e


echoTitle "Building linux_${GOARCH} binaries"
GOOS="linux" GOARCH=$GOARCH  go build -v -o ${APPNAME}_linux_${GOARCH} ./${MAIN_FILE}


echoTitle "Building darwin_${GOARCH} binaries"
GOOS="darwin" GOARCH=$GOARCH  go build -v -o ${APPNAME}_darwin_${GOARCH} ./${MAIN_FILE}


set +e
echoTitle "Done"
