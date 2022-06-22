#!/bin/bash

for x in backend/*/ ; do
    pushd $x
    go get -u ./...
    go mod tidy 
    popd
done