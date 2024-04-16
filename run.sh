#!/bin/bash -x
BINS="HBreq.bin HBrsp.bin SEreq.bin SErsp.bin SMreq.bin SMrsp.bin SDreq.bin SDrsp.bin"
make local && \
bin/merge samples/SEreq.bin samples/SMreq.bin && \
bin/getter samples/ng40_3003_1140/merged.bin && \
pushd samples && ../bin/parser  $BINS && popd
#bin/pcap ~/Downloads/*pcap ~/Downloads/*pcapng ~/workspace/ng40/*pcap   ~/workspace/ng40/*pcapng | grep "validation and reserialisation succeeded" | wc

# # this fails, because the SER is incomplete, even after the merge (the 'ng40')
# bin/getter samples/SEreq.bin && \

#
