package next

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

const (
	// NOTE: the -r flag:
	LS_CMD = `ls -r [0-9]*.md`
	// we only care about the number group:
	FILE_REGEXP   = `(?m)^([0-9]+)\.[[:alpha:]]*\.*md$`
	NOT_FOUND_MSG = "um files not found"
)

var fileRegexp *regexp.Regexp = regexp.MustCompile(FILE_REGEXP)

// calls ls to get the lexical last file
func last() (string, error) {
	// NOTE: globbing requires invoking the shell as cmd:
	cmd := exec.Command("sh", "-c", LS_CMD)

	// fine grained control:
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// the returned err is ExitStatus, which has stderr embedded, but i find this clearer:
	if err := cmd.Run(); err != nil {
		return "", errors.New(stderr.String())
	}

	lines := strings.Split(stdout.String(), "\n")
	if len(lines) == 0 {
		return "", errors.New(NOT_FOUND_MSG)
	}
	return lines[0], nil
}

func Last() {
	s, err := last()
	if err != nil {
		log.Fatalf("um last: %v", err)
	}
	log.Println(s)
}
