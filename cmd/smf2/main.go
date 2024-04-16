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
	"os"
	"sync"

	"pfcpcore/loginit"
	"pfcpcore/smf"

	log "github.com/sirupsen/logrus"
)

var (
	deleteSessions                                         bool
	pauseDuration                                          int64
	nodeId, sgwName, localName, pgwName, upfName, logLevel string
	endFlag                                                sync.Mutex
)

const (
	doIdleTransition     = false
	doMobilityTransition = true
)

func main() {
	flag.BoolVar(&deleteSessions, "delete", false, "delete sessions after creation")
	flag.Int64Var(&pauseDuration, "t", 1000, "millisecond delay between steps")
	flag.StringVar(&pgwName, "pgw", "", "PGWu addr:port")
	flag.StringVar(&sgwName, "sgw", "", "SGWu addr:port")
	flag.StringVar(&upfName, "upf", "", "UPF addr:port")
	flag.StringVar(&localName, "local", "", "local SGWc/PGWc addr:port")
	flag.StringVar(&nodeId, "node", "smf", "PFCP node ID as string")
	flag.StringVar(&logLevel, "loglevel", "debug", "logging level")

	flag.Parse()
	loginit.Init(logLevel)

	if upfName == "" {
		if localAddr, err := netip.ParseAddrPort(localName); err != nil {
			log.Fatalf("failed to parse local endpoint address (%s) as addr/port", localName)
		} else if sgwAddr, err := netip.ParseAddrPort(sgwName); err != nil {
			log.Fatalf("failed to parse sgw endpoint address (%s) as addr/port", sgwName)
		} else if pgwAddr, err := netip.ParseAddrPort(pgwName); err != nil {
			log.Fatalf("failed to parse pgw endpoint address (%s) as addr/port", pgwName)
		} else if sgwAssociation, err := smf.CreateAssociation(localAddr, sgwAddr); err != nil {
			log.Fatal("failed to start sgw association", err)
		} else if pgwAssociation, err := sgwAssociation.Clone(pgwAddr); err != nil {
			log.Fatal("failed to start pgw association", err)
		} else {
			endFlag.Lock()
			go sgw(sgwAssociation, session1)
			pgw(pgwAssociation, session1)
			endFlag.Lock()
		}
	} else if sgwName == "" && pgwName == "" {
		if localAddr, err := netip.ParseAddrPort(localName); err != nil {
			log.Fatalf("failed to parse local endpoint address (%s) as addr/port", localName)
		} else if upfAddr, err := netip.ParseAddrPort(upfName); err != nil {
			log.Fatalf("failed to parse upf endpoint address (%s) as addr/port", upfName)
		} else if upfAssociation, err := smf.CreateAssociation(localAddr, upfAddr); err != nil {
			log.Fatal("failed to start upf association", err)
		} else {
			upf(upfAssociation, session1)
		}
	} else {
		log.Fatal("can only run as SMF or PGWc/SGWc")
	}
}

func sgw(association *smf.Association, session Session) {
	if session, err := association.CreateSession(session.sgwSessionCreate()...); err != nil {
		log.Fatalf("failed to create session")
	} else {

		if doIdleTransition {
			pause("xgwSessionModifyIdle")
			fmt.Fprintf(os.Stderr, "\n\n\n *** xgwSessionModifyIdle ***\n\n")
			session.Modify(session1.xgwSessionModifyIdle()...)

			pause("xgwSessionModifyActive")
			session.Modify(session1.sgwSessionModifyActive()...)

		} else if doMobilityTransition {

			pause("sgwSessionMobilityNewTunnel")
			session.Modify(session1.sgwSessionMobilityNewTunnel()...)

			pause("sgwSessionMobility")
			session.Modify(session1.sgwSessionMobility()...)

			pause("sgwSessionMobilityRemoveTunnel")
			session.Modify(session1.sgwSessionMobilityRemoveTunnel()...)
		}

		if deleteSessions {

			pause("Session Delete")
			session.Delete()
		}

	}
	endFlag.Unlock()
}

func pgw(association *smf.Association, session Session) {
	if session, err := association.CreateSession(session.pgwSessionCreate()...); err != nil {
		log.Fatalf("failed to create session")
	} else if deleteSessions {
		pause()
		session.Delete()
	}
}

func upf(association *smf.Association, session Session) {
	if session, err := association.CreateSession(session.smfSessionCreate()...); err != nil {
		log.Fatalf("failed to create session")
	} else {
		if doIdleTransition {
			pause()
			session.Modify(session1.xgwSessionModifyIdle()...)
			pause()
			session.Modify(session1.smfSessionModifyActive()...)
		}

		if deleteSessions {
			pause()
			session.Delete()
		}
	}
}
