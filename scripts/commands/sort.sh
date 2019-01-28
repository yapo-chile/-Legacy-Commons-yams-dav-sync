
SOURCE="${BASH_SOURCE%/*}"
if [[ ! -d "$SOURCE" ]]; then DIR="$PWD"; fi
. "$SOURCE/colors.sh"

DIR="${IMAGES_PATH}"

START=$(date +%s)

echoHeader "Sorting : $DIR"
echoHeader "Started at $(date +%T)"

if [ $(uname -s) = "Linux" ]
then
    find $DIR -name "*.jpg" -printf '%TY%Tm%TdT%TH%TM%0.2TS %f\n' |sort >  ${YAMS_IMAGES_LIST_FILE}
elif [ $(uname -s) = "Darwin" ]
then
    find $DIR/ |xargs stat  -f "%Sm %N" -t "%Y%m%dT%H%M%S" $DIR/*|sort |sed "s+/\(.*\)${DIR}/+1+g;s+./\(.*\)/+ +g; s+/++g" | grep ".jpeg\|.jpg\|.png\|.gif" >  ${YAMS_IMAGES_LIST_FILE}
fi

END=$(date +%s)
DIFF=$(( $END - $START ))

echoHeader "Ended at $(date +%T)"
echoHeader "Sorting took $DIFF seconds"
echoTitle "Output size      :$(du -h ${YAMS_IMAGES_LIST_FILE})"
echoTitle "Listed images    :$(wc -l ${YAMS_IMAGES_LIST_FILE} | awk '{print $1}')"