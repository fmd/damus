#!/bin/bash
sudo apt-get -y update
sudo apt-get -y install golang docker.io
mkdir go/
export GOPATH=go/
go get .
go build .
