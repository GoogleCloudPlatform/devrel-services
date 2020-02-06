#!/usr/bin/env bash

set -e
echo "" > coverage_integration.txt

dirs=( "drghs-worker" "samplr" "sprvsr" )

for d in "${dirs[@]}"; do
    echo "Testing ./$d/..."
	(
    	cd "./$d"
	    echo "$(pwd)"
	    go test -v \
			    -tags integration \
				-race \
				-coverprofile=profile_integration.out \
				-covermode=atomic \
				./...
	    if [ -f profile_integration.out ]; then
	        cat profile_integration.out >> coverage_integration.txt
	        rm profile_integration.out
	    fi
	)
done
