#!/usr/bin/env bash
export GOPATH=$PWD/../
go build -o helloSrvMac pineal-niwan/grpc_dig/hello/server/
go build -o helloCliMac pineal-niwan/grpc_dig/hello/client/
export GOARCH=amd64
export GOOS=linux
go build -o helloSrvLinux pineal-niwan/grpc_dig/hello/server/
