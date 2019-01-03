# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"


echoTitle "Building bandwith proxy limiter"
set -e

    BUILDDIR="$DIR/../../third-party/floodgate"
    (cd "$BUILDDIR" && ./script/cibuild)
    
set +e
echoTitle "Done"
