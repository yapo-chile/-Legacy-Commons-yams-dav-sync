include scripts/commands/vars.mk

## Install golang system level dependencies
setup:
	@scripts/commands/setup.sh

## Compile the code
build:
	@scripts/commands/build.sh

## Execute the service
run:
	@./${APPNAME}  -command=$(command) -limit=$(limit) -object=$(object) -threads=$(threads)

## Compile and start the service
start: build run

## Display basic service info
info:
	@echo "YO           : ${YO}"
	@echo "ServerRoot   : ${SERVER_ROOT}"
	@echo "API Base URL : ${BASE_URL}"
	@echo "Healthcheck  : curl ${BASE_URL}/api/v1/healthcheck"

 