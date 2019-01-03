package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"uchicago/compiler/lexer"
	"uchicago/compiler/parser"
	"uchicago/compiler/tokenizer"
)

func main() {
	argv := os.Args[1:]

	if len(argv) != 1 || argv[0] == "-h" || argv[0] == "--help" || strings.Contains(argv[0], "help") {
		usage()
		os.Exit(1)
	}

	var buf bytes.Buffer
	bits := strings.Split(filepath.Base(argv[0]), ".") // extract the filename

	// if the filename doesn't actually split into two parts, [<filename>, jack], then we must presume it's a directory
	if len(bits) == 1 {
		root := argv[0]
		argv = argv[:0]

		if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if strings.Contains(path, ".jack") {
				argv = append(argv, path)
			}
			return nil
		}); err != nil {
			log.Fatal(fmt.Sprintf("could not process jack folder (%s): %s\n", root, err.Error()))
		}
	}

	for _, x := range argv {
		bits := strings.Split(filepath.Base(x), ".")
		if len(bits) < 2 || bits[1] != "jack" {
			log.Fatal(fmt.Errorf("incorrect filetype passed: %s", x))
		}

		if err := os.Remove(fmt.Sprintf("%s/%s.xml", filepath.Dir(argv[0]), bits[0])); err != nil {
			//			log.Fatal(fmt.Errorf("could not remove old file: %s", err.Error()))
		}

		if err := os.Remove(fmt.Sprintf("%s/%s.vm", filepath.Dir(argv[0]), bits[0])); err != nil {
			//			log.Fatal(fmt.Errorf("could not remove old file: %s", err.Error()))
		}

		b, vm, err := processFiles(x, bits[0])
		if err != nil {
			log.Fatal(err)
		}

		buf.Write(b)

		if err := ioutil.WriteFile(fmt.Sprintf("%s/%s.xml", filepath.Dir(argv[0]), bits[0]), buf.Bytes(), 0664); err != nil {
			log.Fatal(fmt.Errorf("could not write new file: %s", err.Error()))
		}
		buf.Reset()

		buf.Write(vm)
		if err := ioutil.WriteFile(fmt.Sprintf("%s/%s.vm", filepath.Dir(argv[0]), bits[0]), buf.Bytes(), 0664); err != nil {
			log.Fatal(fmt.Errorf("could not write new vm file: %s", err.Error()))
		}
		buf.Reset()
	}
}

func processFiles(f, fp string) ([]byte, []byte, error) {
	// strip the comments and invalid lines
	m, err := lexer.ProcessFiles(f)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	var out, vm []byte
	b, err := tokenizer.Tokenize(f, fp, m[fp])
	if err != nil {
		return []byte{}, []byte{}, err
	}

	// write out the tokens-only file
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("<tokens>\n"))
	buf.Write(b)
	buf.WriteString(fmt.Sprintf("</tokens>\n"))

	if err := os.Remove(fmt.Sprintf("%s/%sT.xml", filepath.Dir(f), fp)); err != nil {
		//			log.Fatal(fmt.Errorf("could not remove old file: %s", err.Error()))
	}

	if err := ioutil.WriteFile(fmt.Sprintf("%s/%sT.xml", filepath.Dir(f), fp), buf.Bytes(), 0664); err != nil {
		log.Fatal(fmt.Errorf("could not write new file: %s", err.Error()))
	}

	// prepare the parsed, compiled version of tokens
	b, vm, err = parser.Parse(b)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	return append(out, b...), vm, nil
}

// usage prints out the usage for the binary
func usage() {
	fmt.Printf("%s", helpString)
}

var helpString = `Usage: compiler filename|dirname 

	filename|dirname		required filename (single .jack file) or dirname (containing multiple .jack files)  to process
`
