#!/usr/bin/env bash

set -e
rm -rf coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -v -race -covermode=atomic -coverprofile=profile.out $d
    if [ -f profile.out ]; then
        if [ -f coverage.txt ]; then
            cat profile.out |grep -v "mode:" >> coverage.txt
        else
            cat profile.out >> coverage.txt
        fi
        rm profile.out
    fi
done