// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("codegen")
	filePath := "PFCP IEs"
	if len(os.Args) > 1 {
		fmt.Printf("file parameter provided\n")
		filePath = os.Args[1]
	}

	if file, err := os.Open(filePath); err != nil {
		fmt.Printf("can't open: %s (%s)\n", filePath, err.Error())
		os.Exit(1)
	} else {
		var lines []line
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s := scanner.Text()
			fmt.Println(s)
			line := processInputLine(s)
			lines = append(lines, line)
			fmt.Println(line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("error reading file:", err)
			os.Exit(1)
		}
		Codegen(lines[:])
	}
}

type line struct {
	TypeCode int
	TypeName string
}

func (line line) String() string {
	return fmt.Sprintf("%d : '%s'\n", line.TypeCode, line.TypeName)
}

func processInputLine(s string) (l line) {
	before, after, found := strings.Cut(s, "\t")
	if !found {
		fmt.Printf("can't parse: %s\n", s)
		os.Exit(1)
	} else if i, err := strconv.Atoi(before); err != nil {
		fmt.Printf("can't parse: %s\n", s)
		os.Exit(1)
	} else {
		l.TypeCode = i
		l.TypeName = mangle(after)
	}
	return
}

const (
	_0_ rune = 48
	_9_ rune = 57
	_A_ rune = 65
	_Z_ rune = 90
	___ rune = 95
	_a_ rune = 97
	_z_ rune = 122
)

func toUpper(r rune) rune {
	if r >= _a_ && r <= _z_ {
		return r - _a_ + _A_
	} else {
		return r
	}
}

func toLower(r rune) rune {
	if r >= _A_ && r <= _Z_ {
		return r + _a_ - _A_
	} else {
		return r
	}
}

func validIdentifierChar(r rune) bool {
	switch {
	case r == ___:
		return true
	case r >= _0_ && r <= _9_:
		return true
	case r >= _A_ && r <= _Z_:
		return true
	case r >= _a_ && r <= _z_:
		return true
	default:
		return false
	}
}

func mangle(sIn string) (sOut string) {
	capState := true
	for _, c := range sIn {
		if validIdentifierChar(c) {
			if capState {
				capState = false
				sOut = sOut + string(toUpper(c))
			} else {
				sOut = sOut + string(toLower(c))
			}
		} else {
			capState = true
		}
	}
	return
}
