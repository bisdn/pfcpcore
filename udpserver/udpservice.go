// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package udpserver

import (
	"fmt"
	"net"
	"net/netip"

	log "github.com/sirupsen/logrus"
)

const UdpMessageMax = 1500

func Main() {
	fmt.Println("PFCP transport")
}

type UdpServer struct {
	socket          *net.UDPConn
	sendChannel     chan *UdpMessage
	EventChannel    chan UdpEvent
	registeredPeers map[netip.AddrPort]*UdpServerPeer
}

type UdpServerPeer struct {
	Send, Receive chan *UdpMessage
	peer          netip.AddrPort
	parent        *UdpServer
}

func (udpServerPeer *UdpServerPeer) Drop() {
	udpServerPeer.parent.Unregister(udpServerPeer.peer)
}

func (udpServerPeer *UdpServerPeer) Enqueue(udpMessage *UdpMessage) {
	udpServerPeer.Send <- udpMessage
}

func (udpServerPeer *UdpServerPeer) Recirculate(message []byte, port uint16) {
	udpServerPeer.Receive <- &UdpMessage{Payload: message, Remote: port}
}

type UdpMessage struct {
	Payload  []byte
	Remote   uint16
	peerAddr netip.AddrPort
}

type UdpEvent interface {
	isUdpEvent()
}
type UdpEventNetworkError struct{ Err error }

func (UdpEventNetworkError) isUdpEvent() {}

type UdpEventNewPeer struct {
	Payload  []byte
	PeerAddr netip.AddrPort
}

func (UdpEventNewPeer) isUdpEvent() {}

func ToUdpMessage(bytes []byte) *UdpMessage {
	return &UdpMessage{Payload: bytes}
}

func NewUDPServer(local netip.AddrPort) (*UdpServer, error) {
	if socket, err := ListenUDPAddrPort(local); err != nil {
		return nil, err
	} else {
		udpServer := &UdpServer{
			socket:          socket,
			sendChannel:     make(chan *UdpMessage),
			EventChannel:    make(chan UdpEvent),
			registeredPeers: make(map[netip.AddrPort]*UdpServerPeer),
		}
		go udpServer.receive()
		go udpServer.send()
		return udpServer, nil
	}
}

func (udpServer *UdpServer) Drop() {
	udpServer.socket.Close()
}

func (udpServer *UdpServer) receive() {
	for {
		buf := make([]byte, UdpMessageMax)

		if n, addr, err := udpServer.socket.ReadFromUDPAddrPort(buf); err != nil {
			log.Errorf("error in read from port %s\n", err.Error())
			udpServer.EventChannel <- UdpEventNetworkError{err}
			break // TODO review how this is impacting down stream....
			// perhaps the socket should be 'closed' and warnings posted elegantly to peers.....
		} else if peer, ok := udpServer.registeredPeers[addr]; !ok {
			// TODO/Note - 3gpp require that a peer can switch source ports within a session,
			// however in the 'original customer' use case we allow multiple associations from a single source IP,
			// and so cannot tolerate switch of source ports.
			//   But, if that were allowed, this is the location to mange that, by keying on IP, rather than IP:Port
			log.Infof("message from unconfigured peer, read %d from addr %s\n", n, addr)
			udpServer.EventChannel <- UdpEventNewPeer{Payload: buf[:n], PeerAddr: addr}
		} else {
			if peer.Receive != nil {
				peer.Receive <- &UdpMessage{Payload: buf[:n], Remote: addr.Port()}
			}
		}
	}
}

func (udpServer *UdpServer) send() {
	// TODO - handle the cases of error on send
	// But presumably the close down of the channel is meant to be

	for m := range udpServer.sendChannel {
		if _, err := udpServer.socket.WriteToUDPAddrPort(m.Payload, m.peerAddr); err != nil {
			log.Errorf("error in send to port %s\n", err.Error())
			udpServer.EventChannel <- UdpEventNetworkError{err}
		}
	}
	log.Debugf("udpServer send() exits\n")
}

// Note - the only purpose of send worker is to enforce the correct use of source UDP port
// Possibly, it could be replaced by a called function....
func sendWorker(peer netip.AddrPort, in, out chan *UdpMessage) {
	for m := range in {
		if m.Remote == 0 {
			m.peerAddr = peer
		} else {
			m.peerAddr = netip.AddrPortFrom(peer.Addr(), m.Remote)
		}
		out <- m
	}
	log.Debugf("sendWorker(%s) exits\n", peer)
}

func (udpServer *UdpServer) Register(peer netip.AddrPort) (udpServerPeer *UdpServerPeer) {
	send := make(chan *UdpMessage)
	receive := make(chan *UdpMessage)
	udpServerPeer = &UdpServerPeer{
		Send:    send,
		Receive: receive,
		peer:    peer,
		parent:  udpServer,
	}
	udpServer.registeredPeers[peer] = udpServerPeer
	go sendWorker(peer, send, udpServer.sendChannel)
	return
}

func (udpServer *UdpServer) Unregister(peer netip.AddrPort) {
	if udpServerPeer, ok := udpServer.registeredPeers[peer]; !ok {
		log.Errorf("invalid peer as key in Unregister() %s:%s", peer, udpServerPeer.parent.socket.LocalAddr())
	} else {
		delete(udpServer.registeredPeers, peer)
		close(udpServerPeer.Send)
		close(udpServerPeer.Receive)
		udpServerPeer.Receive = nil
		udpServerPeer.Send = nil
	}
}
