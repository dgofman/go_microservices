#!/usr/bin/env bash
#
# This will build for the target operating systems
#
env GOOS=linux GOARCH=amd64 go build -o build/linux/trip-transaction
env GOOS=windows GOARCH=amd64 go build -o build/windows/trip-transaction.exe
env GOOS=darwin go build -o build/mac/trip-transaction