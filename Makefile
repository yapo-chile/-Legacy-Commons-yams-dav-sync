include scripts/commands/vars.mk

## Install golang system level dependencies
setup:
	@scripts/commands/setup.sh

## Compile the code
build:
	@scripts/commands/build.sh

## Run tests and generate quality reports
test:
	@scripts/commands/test.sh

## Run tests and output coverage reports
cover:
	@scripts/commands/test_cover.sh cli

## Run tests and open report on default web browser
coverhtml:
	@scripts/commands/test_cover.sh html

## Run gometalinter and output report as text
checkstyle:
	@scripts/commands/test_style.sh display

## Sort sorts images by date from local dir generating a list dump
sort:
	@scripts/commands/sort.sh

## Execute the service
run:
	@./${APPNAME}  -command=$(command)  -object=$(object) -threads=$(threads)

runsync:
	@./${APPNAME}  -command=sync -dumpfile=${YAMS_IMAGES_LIST_FILE} -threads=$(YAMS_MAX_CONCURRENT_CONN)

removedump:
	rm ${YAMS_IMAGES_LIST_FILE}

## sync starts dav-yams synchronization
sync: build sort runsync removedump

## deleteall the images from yams
deleteall:
 @./${APPNAME}  -command=deleteAll -threads=$(YAMS_MAX_CONCURRENT_CONN)
	
## Compile and start the service
start: build run

## Display basic service info
info:
	@echo "YO           : ${YO}"
	@echo "ServerRoot   : ${SERVER_ROOT}"
	@echo "API Base URL : ${BASE_URL}"
	@echo "Healthcheck  : curl ${BASE_URL}/api/v1/healthcheck"
