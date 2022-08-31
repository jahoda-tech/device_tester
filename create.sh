#!/usr/bin/env bash

./update
name=${PWD##*/}
go get -u all

GOOS=linux GOARCH="amd64" go build -ldflags="-s -w" -o linux/amd64/"$name"
GOOS=linux GOARCH="arm64" go build -ldflags="-s -w" -o linux/arm64/"$name"
GOOS=windows GOARCH="amd64" go build -ldflags="-s -w" -o windows/amd64/"$name"
GOOS=windows GOARCH="arm64" go build -ldflags="-s -w" -o windows/arm64/"$name"
