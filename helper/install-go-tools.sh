#!/bin/bash
GOPATH="$(go env GOPATH)/bin"
PATH="$PATH:$HOME/.local/protoc/bin:$GOPATH"

install_dependencies() {
    PB_REL="https://github.com/protocolbuffers/protobuf/releases"
    VERSION="protoc-3.19.4-linux-x86_64.zip"
    curl -LO $PB_REL/download/v3.19.4/$VERSION

    unzip $VERSION -d $HOME/.local/protoc
    rm $VERSION

    go install github.com/swaggo/swag/cmd/swag@latest
    
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install github.com/golang/mock/mockgen@latest
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest    
}

go_protoc() {
    # shellcheck disable=2035
    protoc \
    -I $HOME/.local/protoc/include -I. \
    --go_out=go \
    --go_opt=paths=source_relative \
    --go-grpc_out=go \
    --go-grpc_opt=paths=source_relative \
    *.proto
}