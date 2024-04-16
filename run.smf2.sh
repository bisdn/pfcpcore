#!/bin/bash -xe
trap "killall sgwpgw upf || :" EXIT
bin/sgwpgw --loglevel=trace --pgw=127.0.0.2:8805 --sgw=127.0.0.3:8805 --upf=127.0.0.4:8805,127.0.0.5:8805 --sgwup=127.0.0.2 &
bin/upf 127.0.0.5:8805 &
bin/smf2 -loglevel trace -local 127.0.0.10:8805 -pgw 127.0.0.2:8805 -sgw 127.0.0.3:8805
sleep 100.0
