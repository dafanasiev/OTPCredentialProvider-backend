#!/usr/bin/env bash
BASEDIR=$(pwd)
PROTOBUF_VERSION=3.4.0
PROTOC_FILENAME=protoc-${PROTOBUF_VERSION}-linux-x86_64.zip

rm -rf $BASEDIR/tmp
mkdir $BASEDIR/tmp
cd $BASEDIR/tmp
wget https://github.com/google/protobuf/releases/download/v${PROTOBUF_VERSION}/${PROTOC_FILENAME}
unzip ${PROTOC_FILENAME}
cd $BASEDIR

go get -u github.com/golang/protobuf/protoc-gen-go

PATH=$PATH:$(go env GOPATH)

$BASEDIR/tmp/bin/protoc --go_out=plugins=grpc:. *.proto
