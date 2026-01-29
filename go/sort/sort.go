package sort

import (
	"errors"
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
	Source flags.String
	Key    flags.String
	Write  flags.Bool
	Help   flags.Bool
}

func initOpts() options {
	return options{
		flags.String{"--source", "-s", "", "path to source list. reads from stdin if omitted."},
		flags.String{"--key", "-k", "", "path to sort key"},
		flags.Bool{"--write", "-w", false, "write sorted list back to --key file"},
		flags.Bool{"--help", "-h", false, "show help"},
	}
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
	opts := initOpts()
	if err := flags.ParseArgs(CMD, SUMMARY, args, &opts); err != nil {
		var herr flags.HelpError
		if errors.As(err, &herr) {
			fmt.Println(herr)
			return
		}
		log.Fatalf("um %s: %s", CMD, err)
	}
	// BORK: by hand for now:
	if !opts.Key.IsSet() {
		fmt.Println(flags.HelpRequired(CMD, opts.Key.Long))
		return
	}

	sslice, err := pipe.FileListSplitMaybeStdin(opts.Source.Val)
	if err != nil {
		log.Fatalf("um %s: %s", CMD, err)
	}
	kslice, err := pipe.FileListSplit(opts.Key.Val)
	if err != nil {
		log.Fatalf("um %s: %s", CMD, err)
	}
	kmap := kMap(kslice)
	out := sort(sslice, kmap)
	if opts.Write.IsSet() {
		write(opts.Key.Val, out)
	} else {
		// to stdout
		fmt.Print(out)
	}
}
