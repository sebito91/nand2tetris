package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	argv := os.Args[1:]
	stripComments := false

	if len(argv) < 1 || argv[0] == "-h" || argv[0] == "--help" || strings.Contains(argv[0], "help") {
		usage()
		os.Exit(1)
	}

	if len(argv) > 1 && argv[1] == "no-comments" {
		fmt.Printf("ERROR -- 'no-comments' provided after filename\n")
		usage()
		os.Exit(1)
	}

	filenames := argv[0:]
	if argv[0] == "no-comments" {
		filenames = argv[1:]
		stripComments = true
	}

	for _, f := range filenames {
		bits := strings.Split(filepath.Base(f), ".")
		if len(bits) < 2 || bits[1] != "in" {
			log.Fatal(fmt.Errorf("incorrect filetype passed: %s", f))
		}
	}

	if err := processFile(stripComments, filenames...); err != nil {
		log.Fatal(err)
	}
}

// usage prints out the usage for the binary
func usage() {
	fmt.Printf("%s", helpString)
}

// stripSpaces maps each character (rune) of the incoming string against a function
// that checks for whitespace. Since we've already removed any blank lines, this function
// simply checks for anything that's considered a utf-8 'space'
func stripSpaces(in string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, in)
}

// processFile takes in a list of filename and strips the file of spaces, writing to filename.out
func processFile(strip bool, filenames ...string) error {
	for _, f := range filenames {
		read, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		var out bytes.Buffer
		for _, x := range strings.Split(fmt.Sprintf("%s", read), "\n") {
			if len(x) == 0 {
				continue
			}

			if strip && strings.HasPrefix(x, "//") {
				continue
			}

			out.WriteString(fmt.Sprintf("%s\n", stripSpaces(x)))
		}

		bits := strings.Split(filepath.Base(f), ".")
		if err := ioutil.WriteFile(fmt.Sprintf("%s/%s.out", filepath.Dir(f), bits[0]), out.Bytes(), 0664); err != nil {
			return err
		}
	}

	return nil
}

var helpString = `Usage: spaces [no-comments] filename <filenames>

	no-comments		an optional parameter to remove all comments from filename
	filenames		optional list of additional filenames (space-delimited)
`
