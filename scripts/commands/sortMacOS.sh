SOURCE="${BASH_SOURCE%/*}"
if [[ ! -d "$SOURCE" ]]; then DIR="$PWD"; fi
. "$SOURCE/colors.sh"


DIR=$IMAGES_PATH
START=$(date +%s)

echoHeader "Sorting : ${IMAGES_PATH}"

find $DIR/ |xargs stat  -f "%Sm %N" -t "%Y%m%dT%H%M%S" $DIR/*|sort |sed "s+/\(.*\)${DIR}/+1+g;s+./\(.*\)/+ +g; s+/++g" | grep ".jpeg\|.jpg\|.png\|.gif" >  ${YAMS_IMAGES_LIST_FILE}

END=$(date +%s)
DIFF=$(( $END - $START ))

echoHeader "Sorting took $DIFF seconds"
echoTitle "Dump file : $YAMS_IMAGES_LIST_FILE"