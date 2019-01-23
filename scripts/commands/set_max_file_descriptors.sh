# Include colors.sh
DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/colors.sh"

set -e

echoTitle "Setting file descriptors"
NEW_MAX=$(($CURRENT_MAX_FILE_DESCRIPTORS+$MAX_FILE_DESCRIPTORS))

sudo ulimit -n ${NEW_MAX}

echoHeader "Max file descriptors: ${NEW_MAX}"
set +e