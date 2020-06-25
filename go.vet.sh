#!/usr/bin/env bash

set -e

dirs=( "devrelservices-admin" "drghs-worker" "leif" "repos" "rtr" "samplr" "sprvsr" )

for d in "${dirs[@]}"; do
    echo "Go vet-ing ./$d/..."
	(
    	cd "./$d"
	    go vet ./...
	)
done

