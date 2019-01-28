#!/usr/bin/env bash

# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"


echoTitle "Building binaries"
set -e

echoTitle "Building linux_${GOARCH} binaries"
GOOS="linux" GOARCH=$GOARCH  go build -v -o ./output/${APPNAME}/${APPNAME}_linux_${GOARCH} ./${MAIN_FILE}

echoTitle "Building darwin_${GOARCH} binaries"
GOOS="darwin" GOARCH=$GOARCH  go build -v -o ./output/${APPNAME}/${APPNAME}_darwin_${GOARCH} ./${MAIN_FILE}

echoTitle "Copying essentials files"

mkdir -p ./output/${APPNAME}/scripts/commands/ && cp -r ./scripts/commands/* ./output/${APPNAME}/scripts/commands/
mkdir -p ./output/${APPNAME}/third-party/ && cp -r ./third-party/* ./output/${APPNAME}/third-party/
mkdir -p ./output/${APPNAME}/migrations/ && cp -r ./migrations/* ./output/${APPNAME}/migrations/
cp ./README.md ./output/${APPNAME}/
cp ./*.rsa ./output/${APPNAME}/
sed "s/build buildbandwidthlimiter/buildbandwidthlimiter/g;s/build run/run/g" Makefile > ./output/${APPNAME}/Makefile

echoTitle "Compressing"

cd ./output/ && tar -czvf ${APPNAME}.tar.gz ${APPNAME}/*

set +e
echoTitle "Done.\nDecompress this file on your server: /output/${APPNAME}.tar.gz "
