package cat

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
	"github.com/brtholomy/um/go/pipe"
)

const (
	CMD       = cmd.Cat
	SUMMARY   = "cat um files together using a filelist"
	SEPARATOR = "\n---\n\n"
)

type options struct {
	Filelist flags.Arg
	Base     flags.Flag[string]
	Help     flags.Flag[bool]
}

func initOpts() options {
	return options{
		flags.Arg{"", "filelist. accepts from stdin if not provided"},
		flags.Flag[string]{"--base", "-b", "", "base directory prepended to files in filelist"},
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

func cat(files []string, base string) (string, error) {
	ff := make([]string, 0, len(files))
	for _, f := range files {
		f := filepath.Join(base, f)
		dat, err := os.ReadFile(f)
		if err != nil {
			return "", fmt.Errorf("error opening target file: %w", err)
		}
		ff = append(ff, string(dat))
	}
	return strings.Join(ff, SEPARATOR), nil
}

func Cat(args []string) {
	opts := initOpts()
	flags.ParseArgs(CMD, SUMMARY, args, &opts)
	files, err := pipe.FileListSplitMaybeStdin(opts.Filelist.Val)
	if err != nil {
		log.Fatalf("um %s: %s", CMD, err)
	}
	s, err := cat(files, opts.Base.Val)
	if err != nil {
		log.Fatalf("um %s: %s", CMD, err)
	}
	fmt.Print(s)
}
