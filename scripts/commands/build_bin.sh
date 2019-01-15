#!/usr/bin/env bash

# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"


echoTitle "Building binaries"
set -e

go build -v -o ./output/${APPNAME}/${APPNAME} ./${MAIN_FILE}

mkdir -p ./output/${APPNAME}/scripts/commands/ && cp -r ./scripts/commands/* ./output/${APPNAME}/scripts/commands/
mkdir -p ./output/${APPNAME}/third-party/ && cp -r ./third-party/* ./output/${APPNAME}/third-party/
mkdir -p ./output/${APPNAME}/migrations/ && cp -r ./migrations/* ./output/${APPNAME}/migrations/
cp ./README.md ./output/${APPNAME}/
cp ./*.rsa ./output/${APPNAME}/
sed "s/ build buildbandwidthlimiter/buildbandwidthlimiter/g" Makefile > ./output/${APPNAME}/Makefile

echoTitle "Compressing"

cd ./output/ && tar -czvf ${APPNAME}.tar.gz ${APPNAME}/*

set +e
echoTitle "Done. File generated. Check: /output/${APPNAME}.tar.gz "
