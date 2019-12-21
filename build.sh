#!/usr/bin/env bash 
set -xe

go get
go build -o bin/application cmd/main.go