#!/bin/bash

set -eux

ORIG_PWD="${PWD}"

# Create our own GOPATH
export GOPATH="${ORIG_PWD}/go"

# Symlink our source dir from inside of our own GOPATH
mkdir -p "${GOPATH}/src/github.com/govau"
ln -s "${ORIG_PWD}/src" "${GOPATH}/src/github.com/govau/cga-image-rotator"

# Build the things
go install github.com/govau/cga-image-rotator

# Copy artefacts to output directory
mkdir -p "${ORIG_PWD}/build"
cp "${GOPATH}/bin/cga-image-rotator" "${ORIG_PWD}/build/cga-image-rotator"
cp "${ORIG_PWD}/src/manifest.yml" "${ORIG_PWD}/build/manifest.yml"
cp -R ${ORIG_PWD}/src/assets/* "${ORIG_PWD}/build/"

# Print it out
find "${ORIG_PWD}/build" -ls
