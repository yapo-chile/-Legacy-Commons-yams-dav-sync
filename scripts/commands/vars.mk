#!/usr/bin/env bash
export UNAMESTR = $(uname)
export GO_FILES = $(shell find . -iname '*.go' -type f | grep -v vendor | grep -v pact) # All the .go files, excluding vendor/ and pact/
GENPORTOFF?=0
genport = $(shell expr ${GENPORTOFF} + \( $(shell id -u) - \( $(shell id -u) / 100 \) \* 100 \) \* 200 + 30100 + $(1))

# BRANCH info from travis
export BUILD_BRANCH=$(shell if [ "${TRAVIS_PULL_REQUEST}" = "false" ]; then echo "${TRAVIS_BRANCH}"; else echo "${TRAVIS_PULL_REQUEST_BRANCH}"; fi)

# REPORT_ARTIFACTS should be in sync with `RegexpFilePathMatcher` in
# `reports-publisher/config.json`
export REPORT_ARTIFACTS=reports

# APP variables
# This variables are for the use of your microservice. This variables must be updated each time you are creating a new microservice
export APPNAME=yams-dav-sync
export YO=`whoami`
export OS:=$(shell uname -s | tr '[:upper:]' '[:lower:]')
export GOARCH=amd64
export SERVICE_HOST=:$(call genport,2)
export SERVER_ROOT=${PWD}
export BASE_URL="http://${SERVICE_HOST}"
export MAIN_FILE=cmd/${APPNAME}/main.go
export LOGGER_SYSLOG_ENABLED=false
export LOGGER_STDLOG_ENABLED=true
export LOGGER_LOG_LEVEL=3

#DATABASE variables
export DATABASE_NAME=postgres
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USER=pgdb
export DATABASE_PASSWORD=postgres
export DATABASE_SSL_MODE=disable
export DATABASE_MAX_IDLE=10
export DATABASE_MAX_OPEN=100
export DATABASE_MIGRATIONS_FOLDER=migrations
export DATABASE_CONN_RETRIES=3

# YAMS variables

# BUCKET LIST FOR DEV:
# POYA 1: 8c2ab775-a9a5-48fb-966f-b1a1b154af13
# POYA 2: b98f66eb-bd6b-47fa-b125-5da03b7534ab
# POYA 3: c637a5e5-1323-42cb-b6ee-7c74de09d6ad
# POYA 4: 5c00e9de-41da-4317-8d2e-03687c57f0dc

export YAMS_MGMT_URL=https://mgmt-us-east-1-yams.schibsted.com/api/v1
export YAMS_TENTAND_ID=f502a79d-9ec7-4778-a580-205223e4d620
export YAMS_DOMAIN_ID=d2b88e84-d868-43b2-af96-456464ba9f5f
export YAMS_BUCKET_ID=8c2ab775-a9a5-48fb-966f-b1a1b154af13
export YAMS_ACCESS_KEY_ID=b73145eec0bd48a2
export YAMS_PRIVATE_KEY=${PWD}/writer-dev.rsa# Your RSA key filepath
export YAMS_IMAGES_LIST_FILE:=dump_images_list.yams# Temp file used to list images to upload
export YAMS_UPLOAD_LIMIT=0
export YAMS_MAX_CONCURRENT_CONN=100# Threads qty used to upload images
export YAMS_TIMEOUT=120
export YAMS_LISTING_LIMIT=0
export YAMS_DELETING_LIMIT=0

# Circuit breaker variables
export CIRCUIT_BREAKER_NAME=HTTP_HANDLER
export CIRCUIT_BREAKER_CONSECUTIVE_FAILURE=10
export CIRCUIT_BREAKER_FAILURE_RATIO=0.6
export CIRCUIT_BREAKER_TIMEOUT=10
export CIRCUIT_BREAKER_INTERVAL=5

# Bandwidth proxy limiter variables
export BANDWIDTH_PROXY_LIMIT=20000# kbps
export BANDWIDTH_PROXY_HOST=localhost:9999
export BANDWIDTH_PROXY_CONN_TYPE=tcp
export BANDWIDTH_PROXY_LATENCY=0
export BANDWIDTH_PROXY_PROCESS_NAME=floodgate

# Metrics exporter variables
export METRICS_PORT=8877

export LAST_SYNC_DEFAULT_DATE=30-12-2015# First execution: skip older images than this date

export ERRORS_MAX_RETRIES_PER_ERROR=3# Skip if the error counter is bigger than this number
export ERRORS_MAX_RESULTS_PER_PAGE=10000# Pagination for error list stored in DB

export IMAGES_PATH=/opt/images/images
