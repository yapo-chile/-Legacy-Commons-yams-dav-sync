
SOURCE="${BASH_SOURCE%/*}"
if [[ ! -d "$SOURCE" ]]; then DIR="$PWD"; fi
. "$SOURCE/colors.sh"

DIR="${IMAGES_PATH}"

START=$(date +%s)

echoHeader "Sorting : $DIR"

if [ $(uname -s) = "Linux" ]
then
    find $DIR -printf '%CY%Cm%CdT%CH%CM%0.2CS %f\n' |sort | grep ".jpg" >  ${YAMS_IMAGES_LIST_FILE}
elif [ $(uname -s) = "Darwin" ]
then
    find $DIR/ |xargs stat  -f "%Sm %N" -t "%Y%m%dT%H%M%S" $DIR/*|sort |sed "s+/\(.*\)${DIR}/+1+g;s+./\(.*\)/+ +g; s+/++g" | grep ".jpeg\|.jpg\|.png\|.gif" >  ${YAMS_IMAGES_LIST_FILE}
fi

END=$(date +%s)
DIFF=$(( $END - $START ))

echoHeader "Sorting took $DIFF seconds"
echoTitle "Output: $(du -h ${YAMS_IMAGES_LIST_FILE})"