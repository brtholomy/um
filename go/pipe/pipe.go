package pipe

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const Newline string = "\n"

func isStdinLoaded() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func GetStdin() ([]string, error) {
	if !isStdinLoaded() {
		return nil, errors.New("stdin not loaded")
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: there's got to be a better way:
	s, _ := strings.CutSuffix(string(data), Newline)
	return strings.Split(s, Newline), nil
}

// reads files from stdin if present, otherwise from the filename:
func FileListSplitMaybeStdin(f string) ([]string, error) {
	filelist, err := GetStdin()
	if err != nil {
		// BORK: don't see a neater way of wrapping these two when we depend on the non-nil in this block:
		// if i just return FileListSplit(), we lose the first err
		filelist, err2 := FileListSplit(f)
		if err2 != nil {
			return nil, fmt.Errorf("%w: %w", err, err2)
		}
		return filelist, nil
	}
	return filelist, nil
}

// opens the given filename and splits into lines:
func FileListSplit(f string) ([]string, error) {
	if f == "" {
		return nil, errors.New("filename is empty")
	}
	dat, err := os.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	return strings.Split(string(dat), Newline), nil
}
