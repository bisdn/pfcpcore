// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var sigChan chan struct{}

func pause(s ...string) {
	if pauseDuration == 0 {
		<-sigChan
	} else {
		select {
		case <-sigChan:
		case <-time.After(time.Millisecond * time.Duration(pauseDuration)):
		}

		if len(s) > 0 {
			fmt.Fprintf(os.Stderr, "\n\n\n *** "+s[0]+" ***\n\n")
		}
	}
}

func init() {
	sigChan = make(chan struct{})
	sigs := make(chan os.Signal, 5)

	signal.Notify(sigs, syscall.SIGUSR1)

	go func() {
		for range sigs {
			sigChan <- struct{}{}
		}
	}()
}
