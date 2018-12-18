#!/usr/bin/env bash

# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"
echo $DIR
set -e

mkdir -p ${REPORT_ARTIFACTS}

COVER_FILE=${REPORT_ARTIFACTS}/cover.out
TMP_COVER_FILE=${REPORT_ARTIFACTS}/cover.out.tmp
COVERAGE_REPORT=${REPORT_ARTIFACTS}/coverage.xml
JUNIT_REPORT=${REPORT_ARTIFACTS}/junit-report.xml

echoHeader "Running Unit Tests"

function run_tests {
    # Get packages list except vendor and pact directories
    packages=$(go list ./... | grep -v vendor | grep -v pact )
    # Create cover output file
    echo "mode: count" > ${COVER_FILE}
    # Test all packages from the list
    for package in ${packages}; do
        echo "" > ${TMP_COVER_FILE}
        go test -v -race -covermode="atomic" -coverprofile=${TMP_COVER_FILE} ${package} || status=$?
        sed '/^mode: atomic$/d' ${TMP_COVER_FILE} >> ${COVER_FILE}
    done
    sed -i.bak '/^$/d' ${COVER_FILE}
    return ${status:-0}
}

# Generate test report
echoTitle "Generating test report"
run_tests | tee /dev/tty | go-junit-report > ${JUNIT_REPORT}; test ${PIPESTATUS[0]} -eq 0 || status=${PIPESTATUS[0]}

if [[ $@ == **cli** ]]; then
    # Print code coverage details
    echoTitle "Printing code coverage details"
    go tool cover -func ${COVER_FILE}
elif [[ $@ == **html** ]]; then
    # Open browser with code coverage details
    echoTitle "Printing code coverage details"
    go tool cover -func ${COVER_FILE}
    echoTitle "Displaying coverage on default browser"
    go tool cover -html ${COVER_FILE}
else
    # Generate coverage report
    echoTitle "Generating coverage report"
    gocov convert ${COVER_FILE} | gocov-xml  > ${COVERAGE_REPORT}; test ${PIPESTATUS[0]} -eq 0 || status=${PIPESTATUS[0]}
fi

echoTitle "Done"
exit ${status:-0}
