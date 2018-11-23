#!/usr/bin/env bash

# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"


echoTitle "Building code"
set -e

go build -v -o ${APPNAME} ./${MAIN_FILE}

set +e
echoTitle "Done"
