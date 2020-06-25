#!/usr/bin/env bash

set -e
echo "" > integration_test_coverage.txt

dirs=( "drghs-worker" "leif" "samplr" "sprvsr" )

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
	        cat profile_integration.out >> integration_test_coverage.txt
	        rm profile_integration.out
	    fi
	)
done
