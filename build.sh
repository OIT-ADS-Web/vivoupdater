#!/bin/sh

cd cmd/vivo_indexer
go build
cd ../../
cd cmd/test_produce
go build
cd ../../