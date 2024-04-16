// Copyright 2024 BISDN GmbH
// This program is free software: you can redistribute it and/or modify it under the terms of the GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.
// You should have received a copy of the GNU Affero General Public License along with this program.
package main

import (
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
)

const (
	templateName = "root_template"
	outputFile   = "generated/ie.go"
	templateFile = "templates/ie.template"
)

func Codegen(table []line) {
	fmt.Println("Codegen")
	compiled_template := template.Must(template.ParseFiles(templateFile))

	var code strings.Builder

	if err := compiled_template.ExecuteTemplate(&code, templateName, table); err != nil {
		fmt.Printf("template execution failed for %s, %s\n", templateName, err)
		return
	} else {
		fmt.Printf("template execution succeeded for %s\n", templateName)
	}

	output := []byte(code.String())

	formatted_output, error := format.Source(output)
	if error == nil {
		output = formatted_output
		fmt.Println("generated code passes syntax analysis - formatted")
	} else {
		fmt.Println("generated code fails syntax analysis - unformatted")
	}

	if os.WriteFile(outputFile, output, 0o644) == nil {
		fmt.Printf(" code written to %s\n", outputFile)
	} else {
		fmt.Printf(" code output to %s could not be written ****\n", outputFile)
	}
}
