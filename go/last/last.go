package last

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"

	cmdpkg "github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
)

const cmd = cmdpkg.Last

const (
	GLOB = "[0-9]*.md"
	// we only care about the number group:
	NOT_FOUND_MSG = "um files not found: " + GLOB
)

type options struct {
	Help flags.Flag[bool]
}

func initOpts() options {
	return options{
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

func parseArgs(args []string) options {
	opts := initOpts()
	for _, arg := range args {
		switch {
		case arg == opts.Help.Long || arg == opts.Help.Short:
			flags.Help(cmd, opts)
		default:
			flags.HelpInvalidArg(cmd, arg)
			flags.Help(cmd, opts)
		}
	}
	return opts
}

// get the lexical last file from GLOB
func GlobLast(glob string) (string, error) {
	// NOTE: filepath.Glob is more reliable than a manual ls call:
	filelist, err := filepath.Glob(glob)
	if err != nil {
		return "", err
	}
	if len(filelist) == 0 {
		return "", errors.New(NOT_FOUND_MSG)
	}
	return filelist[len(filelist)-1], nil
}

func Last(args []string) {
	_ = parseArgs(args)
	s, err := GlobLast(GLOB)
	if err != nil {
		log.Fatalf("um %s: %v", cmdpkg.Last, err)
	}
	// send to stdout, not stderr as is default for log.Print:
	fmt.Println(s)
}
