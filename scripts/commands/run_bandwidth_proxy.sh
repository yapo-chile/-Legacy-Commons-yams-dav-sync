# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"


echoTitle "Running bandwidth limiter proxy"
set -e

if [ $(uname -s) = "Linux" ]
then
    ./third-party/floodgate/build/floodgate_linux_amd64/floodgate  -latency=${BANDWIDTH_PROXY_LATENCY} -listen=${BANDWIDTH_PROXY_HOST} -rate=${BANDWIDTH_PROXY_LIMIT} &
elif [ $(uname -s) = "Darwin" ]
then
    ./third-party/floodgate/build/floodgate_darwin_amd64/floodgate  -latency=${BANDWIDTH_PROXY_LATENCY} -listen=${BANDWIDTH_PROXY_HOST} -rate=${BANDWIDTH_PROXY_LIMIT} &
fi

BANDWIDTH_PROXY_PID=$!

set +e
echoHeader "Bandwidh limiter running with PID $BANDWIDTH_PROXY_PID"
