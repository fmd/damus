#!/bin/bash
sudo apt-get -y update
sudo apt-get -y install golang docker.io
mkdir /tmp/go
export GOPATH=/tmp/go
go get .
go build .
