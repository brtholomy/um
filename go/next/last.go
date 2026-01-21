package next

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"regexp"

	cmdpkg "github.com/brtholomy/um/go/cmd"
)

const (
	GLOB = "[0-9]*.md"
	// we only care about the number group:
	FILE_REGEXP   = `(?m)^([0-9]+)\.[[:alpha:]]*\.*md$`
	NOT_FOUND_MSG = "um files not found: " + GLOB
)

var fileRegexp *regexp.Regexp = regexp.MustCompile(FILE_REGEXP)

// get the lexical last file from GLOB
func last() (string, error) {
	// NOTE: filepath.Glob is more reliable than a manual ls call:
	filelist, err := filepath.Glob(GLOB)
	if err != nil {
		return "", err
	}
	if len(filelist) == 0 {
		return "", errors.New(NOT_FOUND_MSG)
	}
	return filelist[len(filelist)-1], nil
}

func Last() {
	s, err := last()
	if err != nil {
		log.Fatalf("um %s: %v", cmdpkg.Last, err)
	}
	// send to stdout, not stderr as is default for log.Print:
	fmt.Println(s)
}
