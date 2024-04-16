#!/bin/bash -e
# run the simple go unit tests that are available....
for d in pfcpcore/pfcp pfcpcore/transport pfcpcore/endpoint pfcpcore/testcases; do
	if ! go test -count=1 $d; then
		echo "$d *** FAIL ***"
	fi
done
