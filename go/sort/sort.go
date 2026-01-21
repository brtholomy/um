package sort

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
	"github.com/brtholomy/um/go/pipe"
)

const (
	CMD     = cmd.Sort
	SUMMARY = "sort a filelist using a provided key filelist"
)

type options struct {
	Source flags.Flag[string]
	Key    flags.Flag[string]
	Write  flags.Flag[bool]
	Help   flags.Flag[bool]
}

func initOpts() options {
	return options{
		flags.Flag[string]{"--source", "-s", "", "path to source list. reads from stdin if omitted."},
		flags.Flag[string]{"--key", "-k", "", "path to sort key"},
		flags.Flag[bool]{"--write", "-w", false, "write sorted list back to --key file"},
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

func parseArgs(args []string) options {
	opts := initOpts()
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == opts.Source.Long || arg == opts.Source.Short:
			i, opts.Source.Val = flags.ValidateIncrementFetchOrExit(args, i)
		case arg == opts.Key.Long || arg == opts.Key.Short:
			i, opts.Key.Val = flags.ValidateIncrementFetchOrExit(args, i)
		case arg == opts.Write.Long || arg == opts.Write.Short:
			opts.Write.Val = true
		case arg == opts.Help.Long || arg == opts.Help.Short:
			flags.Help(CMD, SUMMARY, opts)
		default:
			flags.HelpInvalidArg(CMD, arg)
			flags.Help(CMD, SUMMARY, opts)
		}
	}
	if !opts.Key.IsSet() {
		flags.HelpRequired(CMD, opts.Key.Long)
		flags.Help(CMD, SUMMARY, opts)
	}
	return opts
}

// reads files from stdin if present, otherwise from the filename:
func fileListSplitMaybeStdin(f string) []string {
	filelist, err := pipe.GetStdin()
	if err != nil {
		filelist = fileListSplit(f)
	}
	return filelist
}

// opens the given filename and splits into lines:
func fileListSplit(f string) []string {
	dat, err := os.ReadFile(f)
	if err != nil {
		// if this fails all is lost. just exit.
		log.Fatalf("um %s: error opening file: %s\n%s", CMD, f, err)
	}
	return strings.Split(string(dat), pipe.Newline)
}

// record the order of the given filelist.
func kMap(kslice []string) map[string]int {
	m := make(map[string]int, 0)
	for i, l := range kslice {
		m[l] = i
	}
	return m
}

// write out the provided source slice while respecting the order provided by kmap
func sort(sslice []string, kmap map[string]int) string {
	oslice := make([]string, max(len(sslice), len(kmap))+1)
	for _, l := range sslice {
		// lines not represented in the key get appended to the end:
		if _, ok := kmap[l]; !ok {
			// we don't care that this might leave empty strings between:
			oslice = append(oslice, l)
			continue
		}
		// NOTE: raison d'etre of this whole thang:
		// if present in the key, respect the order:
		oslice[kmap[l]] = l
	}
	// delete empties and add a final newline
	oslice = slices.DeleteFunc(oslice, func(e string) bool { return e == "" })
	return strings.Join(oslice, pipe.Newline) + pipe.Newline
}

func write(file string, content string) {
	if err := os.WriteFile(file, []byte(content), 0664); err != nil {
		log.Fatalf("um %s: error writing file: %s\n%s", CMD, file, err)
	}
}

func Sort(args []string) {
	opts := parseArgs(args)

	sslice := fileListSplitMaybeStdin(opts.Source.Val)
	kslice := fileListSplit(opts.Key.Val)
	kmap := kMap(kslice)
	out := sort(sslice, kmap)
	if opts.Write.IsSet() {
		write(opts.Key.Val, out)
	} else {
		// to stdout
		fmt.Print(out)
	}
}
