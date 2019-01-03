package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"uchicago/bigvm/lexer"
	"uchicago/bigvm/processor"
)

// handleInit is local var to make sure we've added the necessary logic to start functions
var handleInit bool

// hasSys
var hasSys bool

func main() {
	argv := os.Args[1:]

	if len(argv) < 1 || argv[0] == "-h" || argv[0] == "--help" || strings.Contains(argv[0], "help") {
		usage()
		os.Exit(1)
	}

	for _, x := range argv {
		if strings.Contains(x, "Sys.vm") {
			hasSys = true
		}
	}

	var buf bytes.Buffer
	bits := strings.Split(filepath.Base(argv[0]), ".")
	dirs := strings.Split(filepath.Dir(argv[0]), "/")

	if !handleInit && hasSys {
		if err := os.Remove(fmt.Sprintf("%s/%s.asm", filepath.Dir(argv[0]), dirs[len(dirs)-1])); err != nil {
			//			fmt.Printf("WARN -- %s", err.Error())
		}

		buf.WriteString(fmt.Sprintf("%s\n", processor.Init(bits[0])))
	}

	for _, x := range argv {
		bits := strings.Split(filepath.Base(x), ".")
		if len(bits) < 2 || bits[1] != "vm" {
			log.Fatal(fmt.Errorf("incorrect filetype passed: %s", x))
		}

		b, err := processFiles(x, bits[0])
		if err != nil {
			log.Fatal(err)
		}

		buf.Write(b)
	}

	// add in the endless loop for the end
	buf.WriteString(strings.Join([]string{
		"(END)", // create (END) label
		"@END",  // load the addr for the END label
		"0;JMP", // jump automagically back up, endless loop
		"",
	}, "\n"))

	if err := ioutil.WriteFile(fmt.Sprintf("%s/%s.asm", filepath.Dir(argv[0]), dirs[len(dirs)-1]), buf.Bytes(), 0664); err != nil {
		log.Fatal(err)
	}
}

func processFiles(f, fp string) ([]byte, error) {
	// strip the comments and invalid lines
	m, err := lexer.ProcessFiles(f)
	if err != nil {
		return []byte{}, err
	}

	var out []byte
	b, err := processor.ParseFile(f, fp, m[fp])
	if err != nil {
		return []byte{}, err
	}

	return append(out, b...), nil
}

// usage prints out the usage for the binary
func usage() {
	fmt.Printf("%s", helpString)
}

var helpString = `Usage: vm filename <filenames...>

	filename		required filename to process
	filenames       space-delimited list of additional files to merge into a single asm
`
