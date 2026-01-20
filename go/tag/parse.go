package tag

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/brtholomy/um/go/pipe"
)

const (
	// TODO: source in common with last.go
	GLOB = "[0-9]*.md"

	// The layout string must be a representation of:
	// Jan 2 15:04:05 2006 MST
	// 1   2  3  4  5    6  -7
	DATE_FORMAT = "2006.01.02"

	// ^: YYYY.DD.MM$
	DATE_REGEXP = `(?m)^\: ([\.0-9]+?)$`

	// ^+ tag$
	TAG_REGEXP = `(?m)^\+ (.+)$`
)

// regexp.Compile is by far the most expensive operation of ParseContent:
var dateRegexp *regexp.Regexp = regexp.MustCompile(DATE_REGEXP)
var tagRegexp *regexp.Regexp = regexp.MustCompile(TAG_REGEXP)

type Entry struct {
	filename string
	date     time.Time
	content  string
	tags     []string
}

func parseHeader(content *string) string {
	// returns complete string if not found:
	header, _, _ := strings.Cut(*content, "\n\n")
	return header
}

func parseTags(content *string) (tags []string) {
	res := tagRegexp.FindAllStringSubmatch(*content, -1)
	for i := range res {
		// group submatch is indexed at 1:
		// this shouldn't ever fail if there's a result:
		tags = append(tags, res[i][1])
	}
	return tags
}

func parseDate(content *string) (time.Time, error) {
	res := dateRegexp.FindStringSubmatch(*content)
	if len(res) < 2 {
		return time.Time{}, errors.New("failed to find date string")
	}
	return time.Parse(DATE_FORMAT, res[1])
}

func parseContent(filename string, content *string) Entry {
	base := filepath.Base(filename)
	header := parseHeader(content)
	date, _ := parseDate(&header)
	tags := parseTags(&header)
	return Entry{
		base,
		date,
		*content,
		tags,
	}
}

// reads files from stdin if present, otherwise from the glob pattern:
func getFilelist(glob string) []string {
	filelist, err := pipe.GetStdin()
	// otherwise get from the glob:
	if err != nil {
		filelist, err = filepath.Glob(glob)
		if err != nil {
			log.Fatal(err)
		}
	}
	return filelist
}

// create []Entry representing qualifying files in current directory or from stdin
func entriesGlobOrStdin() []Entry {
	filelist := getFilelist(GLOB)

	// NOTE: size 0, capacity specified:
	entries := make([]Entry, 0, len(filelist))
	for _, f := range filelist {
		dat, err := os.ReadFile(f)
		if err != nil {
			log.Fatal(fmt.Errorf("gag: error opening file: %s\n%w", f, err))
		}
		s := string(dat)
		e := parseContent(f, &s)
		entries = append(entries, e)
	}
	return entries
}

// maps tags to a set of filenames
func makeTagmap(entries []Entry) map[string]Set {
	tagmap := map[string]Set{}
	for _, e := range entries {
		for _, tag := range e.tags {
			// allocate submap if necessary:
			if _, ok := tagmap[tag]; !ok {
				tagmap[tag] = Set{}
			}
			tagmap[tag].Add(e.filename)
		}
	}
	return tagmap
}
