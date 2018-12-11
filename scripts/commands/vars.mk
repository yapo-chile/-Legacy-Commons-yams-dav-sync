#!/usr/bin/env bash
export UNAMESTR = $(uname)
export GO_FILES = $(shell find . -iname '*.go' -type f | grep -v vendor | grep -v pact) # All the .go files, excluding vendor/ and pact/
GENPORTOFF?=0
genport = $(shell expr ${GENPORTOFF} + \( $(shell id -u) - \( $(shell id -u) / 100 \) \* 100 \) \* 200 + 30100 + $(1))

# BRANCH info from travis
export BUILD_BRANCH=$(shell if [ "${TRAVIS_PULL_REQUEST}" = "false" ]; then echo "${TRAVIS_BRANCH}"; else echo "${TRAVIS_PULL_REQUEST_BRANCH}"; fi)


# APP variables
# This variables are for the use of your microservice. This variables must be updated each time you are creating a new microservice
export APPNAME=yams-dav-sync
export YO=`whoami`
export SERVICE_HOST=:$(call genport,2)
export SERVER_ROOT=${PWD}
export BASE_URL="http://${SERVICE_HOST}"
export MAIN_FILE=cmd/${APPNAME}/main.go
export LOGGER_SYSLOG_ENABLED=false
export LOGGER_STDLOG_ENABLED=true
export LOGGER_LOG_LEVEL=1

# YAMS variables

export YAMS_MGMT_URL=https://mgmt-us-east-1-yams.schibsted.com/api/v1
export YAMS_TENTAND_ID=f502a79d-9ec7-4778-a580-205223e4d620
export YAMS_DOMAIN_ID=d2b88e84-d868-43b2-af96-456464ba9f5f
export YAMS_BUCKET_ID=8c2ab775-a9a5-48fb-966f-b1a1b154af13
#POYA 1: b98f66eb-bd6b-47fa-b125-5da03b7534ab
export YAMS_ACCESS_KEY_ID=b73145eec0bd48a2
export YAMS_PRIVATE_KEY=writer-dev.rsa

export YAMS_IMAGES_LIST_FILE=dump_$(shell date -u '+%Y%m%dT%H%M%S').yams

# Redis variables
export REDIS_ADDRESS=10.15.1.221:9379

export IMAGES_PATH=/opt/images/
