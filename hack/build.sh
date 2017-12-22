#!/bin/bash

docker build -t hpk .

go build src/kinfo.go -o bin/kinfo
