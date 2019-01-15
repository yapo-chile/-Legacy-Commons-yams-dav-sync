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
	@./${APPNAME}_${OS}_${GOARCH}  -command=$(command)  -object=$(object) -threads=$(threads)

runsync:
	@./${APPNAME}_${OS}_${GOARCH}  -command=sync -dumpfile=${YAMS_IMAGES_LIST_FILE} -threads=$(YAMS_MAX_CONCURRENT_CONN) -limit=$(YAMS_UPLOAD_LIMIT)

runlist:
	@./${APPNAME}_${OS}_${GOARCH}  -command=list

rundeleteall:
	@./${APPNAME}_${OS}_${GOARCH}  -command=deleteAll -threads=$(YAMS_MAX_CONCURRENT_CONN)


# Build bandwidth proxy limit script
buildbandwidthlimiter:
	@scripts/commands/build_bandwidth_proxy.sh

# Run bandwidth proxy limit script
runbandwidthlimiter:
	@scripts/commands/run_bandwidth_proxy.sh

killbandwidthlimiter:
	pkill ${BANDWIDTH_PROXY_PROCESS_NAME}

removedump:
	rm ${YAMS_IMAGES_LIST_FILE}

## sync starts dav-yams synchronization
sync:
	bash -c "trap 'trap - SIGINT SIGTERM ERR;${MAKE} killbandwidthlimiter removedump; exit 1' SIGINT SIGTERM ERR;${MAKE} trapped-sync"

## list prints objects in yams bucket
list:
	bash -c "trap 'trap - SIGINT SIGTERM ERR;${MAKE} killbandwidthlimiter; exit 1' SIGINT SIGTERM ERR;${MAKE} trapped-list"

## deleteall deletes every image in yams
deleteall:
	bash -c "trap 'trap - SIGINT SIGTERM ERR;${MAKE} killbandwidthlimiter; exit 1' SIGINT SIGTERM ERR;${MAKE} trapped-deleteall"

trapped-sync: build buildbandwidthlimiter runbandwidthlimiter sort runsync killbandwidthlimiter removedump

trapped-list: build buildbandwidthlimiter runbandwidthlimiter runlist killbandwidthlimiter

trapped-deleteall: build buildbandwidthlimiter runbandwidthlimiter rundeleteall killbandwidthlimiter

compress:
	@scripts/commands/compress.sh

## Compile and start the service
start: build run

## Display basic service info
info:
	@echo "YO           : ${YO}"
	@echo "ServerRoot   : ${SERVER_ROOT}"
	@echo "API Base URL : ${BASE_URL}"
	@echo "Healthcheck  : curl ${BASE_URL}/api/v1/healthcheck"
