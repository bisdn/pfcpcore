// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

/*

emulate an SMF for simple test purposes

*/

import (
	"flag"
	"fmt"
	"net/netip"
	"time"

	log "github.com/sirupsen/logrus"

	"pfcpcore/endpoint"
	"pfcpcore/loginit"
	"pfcpcore/pfcp"
	"pfcpcore/utils"
)

var (
	localPort uint

	deleteSessions, overlapped_sending, async_sending, send_nodeID_as_IPv4 bool

	nodeId, sgwName, localName, peerName, logLevel, actions string
	nodeIp                                                  netip.Addr
	recoveryTime                                            uint32
)

const defaultActions = "associate,create,release,remove"

func main() {
	flag.BoolVar(&deleteSessions, "delete", false, "delete sessions after creation")
	flag.BoolVar(&overlapped_sending, "overlapped_sending", false, "send multiple SER before first SMR")
	flag.BoolVar(&async_sending, "async_sending", false, "async_sending ???")
	flag.BoolVar(&send_nodeID_as_IPv4, "send_nodeID_as_IPv4", false, "send_nodeID_as_IPv4???")
	flag.UintVar(&localPort, "port", 58805, "localPort???")
	flag.StringVar(&peerName, "peer", "127.0.0.2:8805", "PFCP peer addr:port")
	flag.StringVar(&sgwName, "sgw", "", "PFCP peer addr:port")
	flag.StringVar(&localName, "local", "", "local PFCP addr:port")
	flag.StringVar(&nodeId, "node", "smf", "PFCP node ID as string")
	flag.StringVar(&logLevel, "loglevel", "debug", "logging level")
	flag.StringVar(&actions, "actions", defaultActions, "customise action set (associate,create,release,remove)")

	flag.Parse()
	loginit.Init(logLevel)

	recoveryTime = pfcp.GetRecoveryTime()

	if sgwName != "" {
		go startEndpoint("pgw", peerName, uint16(localPort))
		startEndpoint("sgw", sgwName, uint16(localPort+1))
	} else {
		startEndpoint("upf", peerName, uint16(localPort))
	}
}

func startEndpoint(role, peerName string, localPort uint16) {
	var peerAddr, localAddr netip.AddrPort
	var err error
	if localName == "" {
		if localAddr, peerAddr, err = utils.GetEndpointAddresses(peerName, localPort); err != nil {
			log.Fatalf("failed to parse %s as IP address - error %s\n", peerName, err)
		} else {
			log.Infof("using default local source addresses, role:%s local:%s peer:%s", role, localAddr, peerAddr)
		}
	} else {
		if localIpAddr, err := netip.ParseAddr(localName); err != nil {
			log.Fatalf("failed to parse localName:%s as IP address - error %s\n", localName, err)
		} else if peerAddr, err = netip.ParseAddrPort(peerName); err != nil {
			log.Fatalf("failed to parse peerName:%s as IP address - error %s\n", peerName, err)
		} else {
			localAddr = netip.AddrPortFrom(localIpAddr, localPort)
			log.Infof("using assigned local source addresses, role:%s local:%s peer:%s", role, localAddr, peerAddr)
		}
	}

	if local, err := endpoint.NewPfcpEndpoint(localAddr); err != nil {
		log.Fatalf("failed to create endpoint for  %s (%s)\n", localAddr, err)
	} else {

		nodeIp = localAddr.Addr()
		peer := local.Peer(peerAddr)
		go heartbeatRunner(peer)

		log.Printf("using local:%s peer:%s for peer role %s\n", localAddr, peerAddr, role)
		var (
			mode  mode
			pause time.Duration
		)
		switch role {

		case "sgw":
			mode = sgwMode
			pause = time.Second * 4

		case "pgw":
			mode = pgwMode
			pause = time.Second * 2

		case "upf":
			mode = upfMode

		default:
			panic("unknown mode")

		}

		var sessionRequests []*pfcp.PfcpMessage
		if role == "pgw" {
			sessionRequests = []*pfcp.PfcpMessage{mode.sessionCreate(&ue_session_1)}
		} else {
			sessionRequests = mode.sessionCreateWithModify(&ue_session_1)
		}

		doRequest(peer, associationRequest())
		if seid, _ := doSessionEstablishmentRequests(peer, sessionRequests...); err != nil {
			log.Errorf("doSessionEstablishmentRequests() fail: %s", err.Error())
		} else {
			log.Errorf("doSessionEstablishmentRequests() success: %s", seid)
			if deleteSessions {
				time.Sleep(pause)
				doRequest(peer, sessionDelete(seid))
			}
		}
	}
}

func doSessionEstablishmentRequests(peer *endpoint.PfcpPeer, reqs ...*pfcp.PfcpMessage) (pfcp.SEID, error) {
	if len(reqs) == 0 {
		return 0, fmt.Errorf("invalid usage of doSessionEstablishmentRequest")
	} else if response, err := peer.BlockingRequest(reqs[0]); err != nil {
		return 0, err
	} else if cause, err := response.Node().ReadCauseCode(); err != nil {
		return 0, err
	} else if cause != pfcp.CauseAccepted {
		return 0, fmt.Errorf("sessionRequest request failed - peer reject with cause %d", cause)
	} else if fseid, err := response.Node().Getter().GetByTc(pfcp.F_SEID).DeserialiseFSeid(); err != nil {
		// // TODO workout why this did not fail!!!!
		// } else if seid, err := response.Node().Getter().GetByTc(pfcp.F_SEID).DeserialiseU64(); err != nil {
		return 0, err
	} else {
		seid := pfcp.SEID(fseid.Seid)
		log.Debugf(("sessionRequest success, seid: %s"), seid)
		for _, msg := range reqs[1:] {
			msg.SEID = &seid
			doRequest(peer, msg)
		}
		return seid, nil
	}
}

func doRequest(peer *endpoint.PfcpPeer, request *pfcp.PfcpMessage) {
	if response, err := peer.BlockingRequest(request); err != nil {
		log.Fatalf(("%s request failed - transport failure"), request.MessageTypeCode)
	} else if cause, err := response.Node().ReadCauseCode(); err == nil {
		if cause == pfcp.CauseAccepted {
			log.Printf(("%s request success"), request.MessageTypeCode)
		} else {
			log.Fatalf(("%s request failed with cause %d\n"), request.MessageTypeCode, cause)
		}
	} else if request.MessageTypeCode == pfcp.PFCP_Heartbeat_Response {
		log.Printf(("%s request success"), request.MessageTypeCode)
	} else {
		log.Fatalf(("%s request failed, missing cause code\n"), request.MessageTypeCode)
	}
}

// func doRequests(peer *endpoint.PfcpPeer, requests ...*pfcp.PfcpMessage) {
// 	for _, request := range requests {
// 		if response, err := peer.BlockingRequest(request); err != nil {
// 			log.Fatalf(("%s request failed - transport failure"), request.MessageTypeCode)
// 		} else if cause, err := response.Node().ReadCauseCode(); err == nil {
// 			if cause == pfcp.CauseAccepted {
// 				log.Printf(("%s request success"), request.MessageTypeCode)
// 			} else {
// 				log.Fatalf(("%s request failed with cause %d\n"), request.MessageTypeCode, cause)
// 			}
// 		} else if request.MessageTypeCode == pfcp.PFCP_Heartbeat_Response {
// 			log.Printf(("%s request success"), request.MessageTypeCode)
// 		} else {
// 			log.Fatalf(("%s request failed, missing cause code\n"), request.MessageTypeCode)
// 		}
// 	}
// }

// func custom() {

// 	if flag.Args()[0] == "repeat" {
// 		if len(flag.Args()) < 2 {
// 			log.Fatal(("repeat count required"))
// 		} else if i, err := strconv.Atoi(flag.Args()[1]); err != nil {
// 			log.Fatal(("repeat count required"))
// 		} else {
// 			log.Println(("start session repeat"))
// 			doRequest(peer, associationRequest())
// 			session := ue_session_1
// 			if overlapped_sending {
// 				for j := 0; j < i; j++ {
// 					localSession := session
// 					session.next()
// 					go doRequest(peer, localSession.partialUpfSER())
// 				}
// 				time.Sleep(time.Second * 10)

// 				session = ue_session_1 // reset the base for second round
// 				for j := 0; j < i; j++ {
// 					localSession := session
// 					session.next()
// 					doRequest(peer, localSession.upfSMR())
// 				}
// 			} else if async_sending {
// 				for j := 0; j < i; j++ {
// 					localSession := session
// 					session.next()
// 					go func() {
// 						doRequest(peer, localSession.partialUpfSER())
// 						doRequest(peer, localSession.upfSMR())
// 					}()
// 				}
// 			} else {
// 				for j := 0; j < i; j++ {
// 					doRequest(peer, session.partialUpfSER())
// 					doRequest(peer, session.upfSMR())
// 					session.next()
// 				}
// 			}

// 			log.Println(("end session repeat"))

// 			time.Sleep(time.Second * 1000)
// 			log.Println(("exit session repeat"))
// 		}
// 	} else {
// 		for _, cmd := range flag.Args()[:] {
// 			switch cmd {
// 			case "a", "ass", "associate":
// 				doRequest(peer, associationRequest())
// 			case "r", "rel", "release":
// 				doRequest(peer, associationReleaseRequest())
// 			case "h", "hb", "heartbeat":
// 				doRequest(peer, heartBeatRequest())
// 			case "d", "del", "delete":
// 				doRequest(peer, ue_session_1.sessionDelete())
// 			case "c", "comp", "complete":
// 				doRequest(peer, ue_session_1.completeUpfSER())
// 			case "p", "part", "partial":
// 				doRequest(peer, ue_session_1.partialUpfSER())
// 			case "s", "smr":
// 				doRequest(peer, ue_session_1.upfSMR())
// 			case "sl", "sleep", "w":
// 				// sleep happens anyway....
// 			}
// 			time.Sleep(time.Second)
// 		}
// 	}
// }

func heartbeatRunner(peer *endpoint.PfcpPeer) {
	log.Println(("start heartbeatRunner"))

	for m := range peer.RequestChan {
		switch m.Message.MessageTypeCode {

		case pfcp.PFCP_Heartbeat_Request:
			reply := pfcp.NewNodeMessage(pfcp.PFCP_Heartbeat_Response, pfcp.IE_RecoveryTimeStamp(recoveryTime))
			peer.EnterResponse(reply, m)
			log.Println(("replied to heartbeat request"))

		default:
			fmt.Println(("got unexpected PFCP request"))
		}
	}

	log.Println(("exit heartbeatRunner"))
}
