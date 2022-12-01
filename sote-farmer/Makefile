#!/bin/bash 

# Parameters for Go
export BINARY_NAME=sote-farmer

export GOPATH=$(HOME)/go

all: | clean build

build:
	go mod tidy
	go build -o bin/$(BINARY_NAME)  ./main.go

clean:
	rm -f bin/$(BINARY_NAME)