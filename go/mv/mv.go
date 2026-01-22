package mv

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/brtholomy/um/go/cat"
	"github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
	"github.com/brtholomy/um/go/next"
	"github.com/brtholomy/um/go/pipe"
)

const (
	CMD     = cmd.Mv
	SUMMARY = "rename an um file descriptor field while updating its header"
)

type options struct {
	Filename   flags.Arg
	Descriptor flags.Arg
	Help       flags.Flag[bool]
}

func initOpts() options {
	return options{
		flags.Arg{"", "old filename"},
		flags.Arg{"", "new descriptor"},
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

// generates new filename and its content with updated H1 header
func newNameAndContent(olds string, opts options) (name string, content string, err error) {
	// name and title
	num, err := next.NumFromLast(opts.Filename.Val)
	if err != nil {
		return "", "", err
	}
	name = fmt.Sprintf("%s.%s.md", num, opts.Descriptor.Val)
	title := fmt.Sprintf("# %s", name)

	// content
	head, tail, ok := strings.Cut(olds, pipe.Newline)
	// the file has no newline at all?
	if !ok {
		tail = head
	}
	// there was no title in the original
	if !strings.HasPrefix(head, cat.H1) {
		tail = olds
	}
	return name, fmt.Sprintf("%s%s%s", title, pipe.Newline, tail), nil
}

// effectively does a mv to the new filename, while updating the H1 header to match.
func mv(opts options) error {
	oldb, err := os.ReadFile(opts.Filename.Val)
	if err != nil {
		return err
	}
	olds := string(oldb)

	name, content, err := newNameAndContent(olds, opts)
	if err != nil {
		return err
	}

	// create
	newf, err := os.Create(name)
	if err != nil {
		return err
	}
	defer newf.Close()
	_, err = newf.WriteString(content)
	if err != nil {
		return err
	}

	// destroy
	err = os.Remove(opts.Filename.Val)
	if err != nil {
		return err
	}
	return nil
}

func Mv(args []string) {
	opts := initOpts()
	flags.ParseArgs(CMD, SUMMARY, args, &opts)
	err := mv(opts)
	if err != nil {
		log.Fatalf("um %s: %v", CMD, err)
	}
}
