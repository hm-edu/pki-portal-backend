#!/bin/bash

for x in backend/*/ ; do
    pushd $x
    OLD=$(cat go.mod | grep "portal-common v0.0.0")
    go get -u ./...
    sed -i 's@^.*portal-common v0.0.0.*$@'"$OLD"'@' go.mod
    go mod tidy 
    popd
done