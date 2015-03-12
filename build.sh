#!/bin/bash

set +e

export GOPATH=/storage/kubernetes-deps-git/

go build
./generate-bash
