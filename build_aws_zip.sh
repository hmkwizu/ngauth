#!/usr/bin/env bash 
set -xe 

# build binary
GOARCH=amd64 GOOS=linux go build -o bin/application cmd/main.go

# build lamda
#GOARCH=amd64 GOOS=linux go build -o bin/generate_otp  cmd/lamda/generate_otp/main.go

# create zip containing the bin folder
zip -r ngauth.zip bin
mv ngauth.zip ../ngauth.zip