package cat

import (
	"fmt"
	"log"
	"os"
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
	Help     flags.Flag[bool]
}

func initOpts() options {
	return options{
		flags.Arg{"", "filelist. accepts from stdin if not provided"},
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

func cat(files []string) (string, error) {
	var total int64
	for _, f := range files {
		stat, err := os.Stat(f)
		if err != nil {
			return "", fmt.Errorf("stat failed: %w", err)
		}
		if !stat.IsDir() && stat.Mode().IsRegular() {
			total += stat.Size()
		}
	}
	ff := make([]string, 0, total)
	for _, f := range files {
		dat, err := os.ReadFile(f)
		if err != nil {
			return "", fmt.Errorf("error opening file: %w", err)
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
	s, err := cat(files)
	if err != nil {
		log.Fatalf("um %s: %s", CMD, err)
	}
	fmt.Print(s)
}
