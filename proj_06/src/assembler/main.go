package main

import (
	"assembler/lexer"
	"assembler/parser"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	argv := os.Args[1:]

	if len(argv) < 1 || argv[0] == "-h" || argv[0] == "--help" || strings.Contains(argv[0], "help") {
		usage()
		os.Exit(1)
	}

	if err := processFiles(argv[0:]...); err != nil {
		log.Fatal(err)
	}
}

func processFiles(filenames ...string) error {
	for _, f := range filenames {
		bits := strings.Split(filepath.Base(f), ".")
		if len(bits) < 2 || bits[1] != "asm" {
			return fmt.Errorf("incorrect filetype passed: %s", f)
		}
	}

	// strip the comments and invalid lines
	m, err := lexer.ProcessFiles(filenames...)
	if err != nil {
		return err
	}

	// process each line in each file
	for _, f := range filenames {
		bits := strings.Split(filepath.Base(f), ".")

		out, err := parser.ParseFile(bits[0], m[bits[0]])
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(fmt.Sprintf("%s/%s.hack", filepath.Dir(f), bits[0]), out, 0664); err != nil {
			return err
		}
	}

	return nil
}

// usage prints out the usage for the binary
func usage() {
	fmt.Printf("%s", helpString)
}

var helpString = `Usage: assembler filename <filenames>

	filenames		optional list of additional filenames (space-delimited)
`
