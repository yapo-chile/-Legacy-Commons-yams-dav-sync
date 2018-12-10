include scripts/commands/vars.mk

## Install golang system level dependencies
setup:
	@scripts/commands/setup.sh

## Compile the code
build:
	@scripts/commands/build.sh

## Sort sorts images generatin yams file
sort:
	@scripts/commands/sortGNU.sh

## Sort sorts images generatin yams file
sortmac:
	@scripts/commands/sortMacOS.sh

## Execute the service
run:
	@./${APPNAME}  -command=$(command)  -object=$(object) -threads=$(threads)

runsync:
	@./${APPNAME}  -command=sync -dumpfile=${YAMS_IMAGES_LIST_FILE} -threads=100

## sync starts dav-yams synchornization using macOS
syncmac: build sortmac runsync

## sync starts dav-yams synchornization using GNU
sync: build sort runsync

## deleteall the images from yams
deleteall:
 @./${APPNAME}  -command=deleteAll -threads=100
	
## Compile and start the service
start: build run

## Display basic service info
info:
	@echo "YO           : ${YO}"
	@echo "ServerRoot   : ${SERVER_ROOT}"
	@echo "API Base URL : ${BASE_URL}"
	@echo "Healthcheck  : curl ${BASE_URL}/api/v1/healthcheck"
