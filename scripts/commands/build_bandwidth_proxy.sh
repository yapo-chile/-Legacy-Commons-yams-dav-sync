# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"
 BUILDDIR="$DIR/../../third-party/floodgate"


set -e
    if [ ! -f "$BUILDDIR/build/floodgate_linux_amd64/floodgate" ] ||  [ ! -f "$BUILDDIR/build/floodgate_darwin_amd64/floodgate" ]; then
        echoTitle "Building bandwidth limiter.."
        (cd "$BUILDDIR" && ./script/cibuild)
        echoTitle "Done"
    fi
set +e
