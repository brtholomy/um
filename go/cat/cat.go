package cat

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
	"github.com/brtholomy/um/go/pipe"
)

const (
	CMD            = cmd.Cat
	SUMMARY        = "cat um files together using a filelist. removes header by default"
	HR_BLOCK       = "\n---\n\n"
	HR_BLOCK_STRIP = "---\n\n"
	DOUBLE_NEWLINE = "\n\n"
	H1             = "# "
)

// NOTE: would prefer to reuse next.FILE_REGEXP, but this is clearer:
// NOTE: supports multiple filenames with a single newline between:
const FILE_LINK_REGEXP = `(?m)` + HR_BLOCK_STRIP + `([0-9]+\.[^\.]*\.*md\n)+\n`

var fileLinkRegexp *regexp.Regexp = regexp.MustCompile(FILE_LINK_REGEXP)

type options struct {
	Filelist       flags.Glob
	Base           flags.String
	KeepHeader     flags.Bool
	KeepTitle      flags.Bool
	StripFileLinks flags.Bool
	Help           flags.Bool
}

func initOpts() options {
	return options{
		flags.Glob{nil, ".um filelist. accepts multiple. reads from stdin if not provided"},
		flags.String{"--base", "-b", "", "base directory prepended to files in filelist"},
		flags.Bool{"--keep-header", "-d", false, "preserve um headers in concatenated file. overrides --keep-title"},
		flags.Bool{"--keep-title", "-t", false, "preserve um titles in concatenated file"},
		flags.Bool{"--strip-file-links", "-s", false, "strip file links in concatenated file"},
		flags.Bool{"--help", "-h", false, "show help"},
	}
}

// remove the um header:
//
// # title
// : date
// + tag
//
// optionally keep just the # title
func decapitate(s string, opts options) string {
	// if there's no header at all, forget it:
	if opts.KeepHeader.IsSet() || !strings.HasPrefix(s, H1) {
		return s
	}
	// I'd rather slice and dice than mess with regex
	head, tail, ok := strings.Cut(s, DOUBLE_NEWLINE)
	if !ok {
		// head is complete if sep wasn't found:
		return head
	}
	s = tail
	if opts.KeepTitle.IsSet() {
		// we just take the first line. but when the header consists only of the title, there is no newline:
		if title, _, ok := strings.Cut(head, pipe.Newline); ok || !strings.Contains(head, pipe.Newline) {
			s = fmt.Sprintf("%s%s%s", title, DOUBLE_NEWLINE, s)
		}
	}
	return s
}

// strips out file links, which are simply a filename per line, between hr section blocks:
// ---
//
// 100.foo.md
// 200.bar.md
//
// NOTE: lists must contain a single newline as separator.
func stripFileLinks(s string, opts options) string {
	if !opts.StripFileLinks.IsSet() {
		return s
	}
	return fileLinkRegexp.ReplaceAllString(s, "")
}

func cat(files []string, opts options) (string, error) {
	ff := make([]string, 0, len(files))
	for _, f := range files {
		bf := filepath.Join(opts.Base.Val, f)
		dat, err := os.ReadFile(bf)
		if err != nil {
			return "", fmt.Errorf("error opening target file: %w", err)
		}
		s := decapitate(string(dat), opts)
		ff = append(ff, s)
	}
	// NOTE: prepend leading HR_BLOCK, since these are used for section numbering in both online and print format:
	catted := HR_BLOCK + strings.Join(ff, HR_BLOCK)
	// NOTE: strip here, because only the fully catted string will match the file link signature,
	// since such links can occur at the beginning of a file with no leading hr.
	catted = stripFileLinks(catted, opts)
	return catted, nil
}

func Cat(args []string) {
	opts := initOpts()
	help := flags.NewHelpError(CMD, SUMMARY)
	if err := flags.ParseArgs(help, args, &opts); err != nil {
		if errors.As(err, &help) {
			fmt.Println(help)
			return
		}
		log.Fatalf("um %s: %s", CMD, err)
	}
	// NOTE: um cat expects .um files, the content of which is assembled:
	files, err := pipe.FileListFromGlobOrStdin(opts.Filelist.Val)
	if err != nil {
		log.Fatalf("um %s: %s", CMD, err)
	}
	s, err := cat(files, opts)
	if err != nil {
		log.Fatalf("um %s: %s", CMD, err)
	}
	fmt.Print(s)
}
