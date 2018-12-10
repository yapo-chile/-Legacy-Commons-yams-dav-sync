#!/usr/bin/env bash

# Disable colors
ENABLE_COLORS=true

# Echo color variables
## Header Color
HC='\033[1;33m'
## Title Color
TC='\033[0;33m'
## Error Color
EC='\033[0;31m'
## No Color
NC='\033[0m'

echoHeader() {
    if [ -z "$1" ]
    then
        return
    fi

    if [ -z "$ENABLE_COLORS" ] || [ "$ENABLE_COLORS" != "true"  ]
    then
        echo "${1}"
    else
        echo "${HC}${1}${NC}"
    fi
}

echoTitle() {
    if [ -z "$1" ]                           
    then
        return
    fi

    if [ -z "$ENABLE_COLORS" ] || [ "$ENABLE_COLORS" != "true"  ]
    then
        echo "${1}"
    else
        echo "${TC}${1}${NC}"
    fi
}

echoError() {
    if [ -z "$1" ]                           
    then
        return
    fi

    if [ -z "$ENABLE_COLORS" ] || [ "$ENABLE_COLORS" != "true"  ]
    then
        echo "${1}"
    else
        echo -e "${EC}${1}${NC}"
    fi
}