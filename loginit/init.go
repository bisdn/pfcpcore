// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package loginit

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func Init(levelString string) {
	if logLevel, err := log.ParseLevel(levelString); err != nil {
		fmt.Fprintf(os.Stderr, "invalid log level: - %s - %s\n", levelString, err.Error())
		os.Exit(1)
	} else {
		log.SetLevel(logLevel)
		log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000", FullTimestamp: true})
	}
}
