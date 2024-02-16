#!/bin/bash

GOLANGCI="${GOPATH}/bin/golangci-lint"
opt="run"

$GOLANGCI $opt ./...
