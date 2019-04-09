#!/bin/sh

cd cmd/vivo_indexer
go build
cd ../../
cd cmd/fake_produce
go build
cd ../../