#!/bin/bash

DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# install main dependencies
echo "Installing additional dependencies…"

cd ${DIR}/../transfer
go get ./...

cd ${DIR}/../hts
go get ./...

cd ${DIR}/../hcs
go get ./...
