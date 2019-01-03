package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"vm/lexer"
	"vm/processor"
)

func main() {
	argv := os.Args[1:]

	if len(argv) < 1 || argv[0] == "-h" || argv[0] == "--help" || strings.Contains(argv[0], "help") {
		usage()
		os.Exit(1)
	}

	if len(argv) > 1 {
		fmt.Printf("[WARN] only taking first element: %s\n", argv[0])
	}

	if err := processFiles(argv[0]); err != nil {
		log.Fatal(err)
	}
}

func processFiles(f string) error {
	bits := strings.Split(filepath.Base(f), ".")
	if len(bits) < 2 || bits[1] != "vm" {
		return fmt.Errorf("incorrect filetype passed: %s", f)
	}

	// strip the comments and invalid lines
	m, err := lexer.ProcessFiles(f)
	if err != nil {
		return err
	}

	out, err := processor.ParseFile(bits[0], m[bits[0]])
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(fmt.Sprintf("%s/%s.asm", filepath.Dir(f), bits[0]), out, 0664); err != nil {
		return err
	}

	return nil
}

// usage prints out the usage for the binary
func usage() {
	fmt.Printf("%s", helpString)
}

var helpString = `Usage: vm filename

	filename		required filename to process
`
