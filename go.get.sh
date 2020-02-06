#!/usr/bin/env bash

set -e
echo "" > coverage.txt

dirs=( "devrelservices-admin" "drghs-worker" "repos" "rtr" "samplr" "sprvsr" )

for d in "${dirs[@]}"; do
    echo "Go getting ./$d/..."
	(
    	cd "./$d"
	    go get ./...
	)
done

go get -u golang.org/x/lint/golint
