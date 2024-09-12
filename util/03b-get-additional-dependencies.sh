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
