#!/usr/bin/env bash
export $(cat ./.env | grep -v ^# | xargs) 
go run ./cmd/vivo_indexer/main.go

