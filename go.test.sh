#!/usr/bin/env bash

set -e
echo "" > coverage.txt

dirs=( "drghs-worker" "samplr" "sprvsr" )

for d in "${dirs[@]}"; do
    echo "Testing ./$d/..."
	(
    	cd "./$d"
	    echo "$(pwd)"
	    go test -v -race -coverprofile=../profile.out -covermode=atomic ./...
	    if [ -f ../profile.out ]; then
	        cat ../profile.out >> ../coverage.txt
	        rm ../profile.out
	    fi
	)
done
