// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package utils

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type numberedPacket struct {
	Number int
	Bytes  []byte
}

func OpenPfcpPcapFile(fp string) ([]numberedPacket, error) {
	var (
		payloads     []numberedPacket
		packetNumber int
	)

	if handle, err := pcap.OpenOffline(fp); err != nil {
		return nil, fmt.Errorf("could not open PCAP file %s (%s)", fp, err.Error())
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			packetNumber += 1
			if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer == nil {
				// packet is not UDP
			} else if udp, _ := udpLayer.(*layers.UDP); udp.DstPort != 8805 {
				// packet is not PFCP
			} else if ap := packet.ApplicationLayer(); ap == nil {
				// packet has no application data
			} else if payload := ap.Payload(); payload == nil {
				// packet has no payload
			} else {
				payloads = append(payloads, numberedPacket{packetNumber, payload})
			}
		}
	}
	return payloads, nil
}
