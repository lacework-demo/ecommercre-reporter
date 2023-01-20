#!/bin/bash

set -e
function log() {
    echo "---> ${1}"
}

## Create output dir
log "creating output dir"
mkdir dist

## Check for required values
log "checking for required values"
if [ ! -d frontend ]; then
    log "must run from root of repo"
    exit 1
fi

## Setup packages
log "setup packages"
(cd frontend && npm i)

## Build Frontend SPA
log "building frontend"
(cd frontend && BUILD_PATH="../dist/frontend" npm run build)

## Build Backend
log "building backend"
make test
make build-cli-cross-platform

log "packaging"
tar --exclude '*.tgz' -zcvf dist/frontend-client.tgz -C ./dist  frontend/

log "success!"
