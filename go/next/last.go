package next

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const (
	// NOTE: the -r flag:
	LS_CMD = `ls -r [0-9]*.md`
	// we only care about the number group:
	FILE_REGEXP = `(?m)^([0-9]+)\.[[:alpha:]]*\.*md$`
)

var fileRegexp *regexp.Regexp = regexp.MustCompile(FILE_REGEXP)

// calls ls to get the lexical last file
func last() (string, error) {
	// NOTE: globbing requires invoking the shell as cmd:
	cmd := exec.Command("sh", "-c", LS_CMD)
	ls, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(ls), "\n")
	if len(lines) == 0 {
		return "", errors.New("no um files in current dir")
	}
	return lines[0], nil
}

func Last() {
	s, err := last()
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	fmt.Println(s)
}
