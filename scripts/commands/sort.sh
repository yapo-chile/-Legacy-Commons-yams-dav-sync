
SOURCE="${BASH_SOURCE%/*}"
if [[ ! -d "$SOURCE" ]]; then DIR="$PWD"; fi
. "$SOURCE/colors.sh"

DIR="${IMAGES_PATH}"

START=$(date +%s)
echoHeader "Sorting started : $(START)"

echoHeader "Sorting : $DIR"

find $DIR -printf '%CY%Cm%CdT%CH%CM%0.2CS %f\n' |sort | grep ".jpg" >  ${YAMS_IMAGES_LIST_FILE}

END=$(date +%s)
DIFF=$(( $END - $START ))

echoHeader "Sorting took $DIFF seconds"
echoTitle "Dump file : $YAMS_IMAGES_LIST_FILE"