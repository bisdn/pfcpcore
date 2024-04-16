// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package utils

import (
	"fmt"
	"os"
)

func GetFileMessages(args []string) (contents [][]byte) {
	if len(args) > 0 {
		for _, filePath := range args {
			if _, err := os.Stat(filePath); err != nil {
				fmt.Printf("Could not read from %s - %s\n", filePath, err.Error())
				os.Exit(1)
			} else if content, err := os.ReadFile(filePath); err != nil {
				fmt.Printf("Could not read from %s - %s\n", filePath, err.Error())
				os.Exit(1)
			} else {
				fmt.Printf("reading %s\n", filePath)
				contents = append(contents, content)
			}
		}
	} else {
		fmt.Printf("no file name supplied\n")
		os.Exit(1)
	}
	return contents
}

func GetArgFileMessages() (contents [][]byte) {
	args := os.Args[1:]
	return GetFileMessages(args)
}

func WriteFileMessage(filePath string, contents []byte) error {
	if err := os.WriteFile(filePath, contents, 0o644); err != nil {
		fmt.Printf("Could not write to %s - %s\n", filePath, err.Error())
		os.Exit(1)
		return err
	}
	return nil
}
