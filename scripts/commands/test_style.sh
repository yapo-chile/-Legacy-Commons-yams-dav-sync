#!/bin/bash

# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"

set -e

mkdir -p ${REPORT_ARTIFACTS}

CHECKSTYLE_FILE=${REPORT_ARTIFACTS}/checkstyle-report.xml

echoHeader "Running Checkstyle Tests"

COMMAND='gometalinter.v2 ./... --config ".gometalinter.json"'
if [[ $@ == **display** ]]; then
    COMMAND="${COMMAND}"
else
    COMMAND="${COMMAND} --checkstyle | tee /dev/tty > ${CHECKSTYLE_FILE}"
fi

eval ${COMMAND}
status=${PIPESTATUS[0]}

# We need to catch error codes that are bigger then 2,
# they signal that gometalinter exited because of underlying error.
if [ ${status} -ge 2 ]; then
    echo "gometalinter exited with code ${status}, check gometalinter errors"
    exit ${status}
fi

exit 0
