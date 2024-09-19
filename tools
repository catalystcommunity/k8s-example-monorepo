#!/usr/bin/env bash

# If you are used to a Makefile, this is where you do the work instead
set -e

THISSCRIPT=$(basename $0)

PROJECT_ROOT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# colors
GREEN=$'\e[0;32m'
NC=$'\e[0m'

# Modify for the help message
usage() {
    echo "${THISSCRIPT} command"
    echo ""
    echo "Commands:"
    echo ""
    echo "  setup       Set things up, obviously"
    echo ""
    exit 0
}

setup(){
    echo "This is where we'll set things up if needs be" 
}

log_status() {
  echo "${GREEN}---------------------------------  ${1}  ---------------------------------${NC}"
}

# This should be last in the script, all other functions are named beforehand.
case "$1" in
    "setup")
        shift
        setup "$@"
        ;;
    *)
        usage
        ;;
esac

exit 0
