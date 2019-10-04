#!/bin/sh

cd ${GITHUB_WORKSPACE:-.}

go test -tags=unit ./...
