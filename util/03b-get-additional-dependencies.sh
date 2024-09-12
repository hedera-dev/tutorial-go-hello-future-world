#!/bin/bash

DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# install main dependencies
echo "Installing additional dependencies…"

cd ${DIR}/../transfer
go get ./...
go mod download

cd ${DIR}/../hts
go get ./...
go mod download

cd ${DIR}/../hcs
go get ./...
go mod download

cd ${DIR}/../transfer
echo "Build 'transfer'"
time go build

cd ${DIR}/../hts
echo "Build 'hts'"
time go build

cd ${DIR}/../hcs
echo "Build 'hcs'"
time go build
