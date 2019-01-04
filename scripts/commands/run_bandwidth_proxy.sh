# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"

set -e

echoTitle "Running bandwidth limiter proxy"

if [ $(uname -s) = "Linux" ]
then
    ./third-party/floodgate/build/floodgate_linux_amd64/floodgate  -latency=${BANDWIDTH_PROXY_LATENCY} -listen=${BANDWIDTH_PROXY_HOST} -rate=${BANDWIDTH_PROXY_LIMIT} &
elif [ $(uname -s) = "Darwin" ]
then
    ./third-party/floodgate/build/floodgate_darwin_amd64/floodgate  -latency=${BANDWIDTH_PROXY_LATENCY} -listen=${BANDWIDTH_PROXY_HOST} -rate=${BANDWIDTH_PROXY_LIMIT} &
fi

BANDWIDTH_PROXY_PID=$!

echoHeader "Bandwidth limiter proxy running with PID $BANDWIDTH_PROXY_PID ($BANDWIDTH_PROXY_PROCESS_NAME)"
set +e
