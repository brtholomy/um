package tag

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	cmdpkg "github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
	"github.com/brtholomy/um/go/pipe"
)

// just so that copied calls into flags.Help don't have to be adjusted between files:
const cmd = cmdpkg.Tag

type options struct {
	Query   flags.Arg
	Date    flags.Flag[string]
	Invert  flags.Flag[bool]
	Verbose flags.Flag[bool]
	Help    flags.Flag[bool]
}

func initOpts() options {
	return options{
		flags.Arg{"", "tag query"},
		flags.Flag[string]{"--date", "-d", "", "date range in ISO 8601: YYYY.MM.DD[-YYYY.MM.DD]"},
		flags.Flag[bool]{"--invert", "-i", false, "invert match"},
		flags.Flag[bool]{"--verbose", "-v", false, "print a verbose summary"},
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

func parseArgs(args []string) options {
	opts := initOpts()
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case i == 0 && !flags.HasDashPrefix(arg):
			opts.Query.Val = arg
		case arg == opts.Invert.Long || arg == opts.Invert.Short:
			opts.Invert.Val = true
		case arg == opts.Verbose.Long || arg == opts.Verbose.Short:
			opts.Verbose.Val = true
		case arg == opts.Date.Long || arg == opts.Date.Short:
			i, opts.Date.Val = flags.ValidateIncrementFetchOrExit(args, i)
		case arg == opts.Help.Long || arg == opts.Help.Short:
			flags.Help(cmd, opts)
		default:
			flags.HelpInvalidArg(cmd, arg)
			flags.Help(cmd, opts)
		}
	}
	return opts
}

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

type Operator string

const (
	SINGLE Operator = ""
	OR     Operator = ","
	AND    Operator = "+"
	WILD   Operator = "*"
)

type Query struct {
	Op   Operator
	Tags []string
}

type Entry struct {
	filename string
	date     time.Time
	content  string
	tags     []string
}

// TODO: consider using this in tagmap and adjacencies. Currently only for printed results for the
// sake of slices.SortFunc
type TagCount struct {
	name  string
	count int
}

// convenience shorthand for this awkward map type.
type Set map[string]bool

// add members to the "set"
func (s Set) Add(mems ...string) {
	for _, m := range mems {
		s[m] = true
	}
}

// get all members in a slice
func (s Set) Members() []string {
	return slices.Collect(maps.Keys(s))
}

// s ∪ t
func (s Set) Union(tt ...Set) {
	for _, t := range tt {
		s.Add(t.Members()...)
	}
}

// s ∩ t
func (s Set) Intersect(t Set) {
	for m, _ := range s {
		if !t[m] {
			delete(s, m)
		}
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

// TODO: only accepts one kind of syntax at a time
func parseQuery(query string) Query {
	// initialize for the single tag case:
	q := Query{
		Op:   SINGLE,
		Tags: []string{query},
	}
	// count a missing tag as WILD:
	if query == "" {
		// TODO: somewhat abusing this concept for the empty query case:
		q.Op = WILD
		return q
	}
	// NOTE: will match OR first
	ops := []Operator{OR, AND}
	for _, op := range ops {
		if s := strings.Split(query, string(op)); len(s) > 1 {
			q.Op = op
			q.Tags = s
			break
		}

	}
	return q
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

func getEntries(filelist []string) []Entry {
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

// shrinks the entries to only include files within a date range.
func dateRange(entries []Entry, date string) []Entry {
	// deleting from the old slice would be less efficient than appending to a new one:
	ranged := make([]Entry, 0, len(entries))
	from, to := time.Time{}, time.Time{}

	// when there's no range, the first string here will be the input:
	f, t, ok := strings.Cut(date, "-")
	from, _ = time.Parse(DATE_FORMAT, f)
	if ok {
		to, _ = time.Parse(DATE_FORMAT, t)
	} else {
		// use the from date for the case of a single date given:
		to, _ = time.Parse(DATE_FORMAT, f)
	}
	for _, e := range entries {
		if from.Compare(e.date) <= 0 && 0 <= to.Compare(e.date) {
			ranged = append(ranged, e)
		}
	}
	return ranged
}

// produce a Set reduced to the files covered by combined queries
func processQueries(tagmap map[string]Set, query Query) Set {
	set := Set{}
	// sanity check:
	if len(query.Tags) < 1 {
		return set
	}

	// initialize as first query
	q := query.Tags[0]

	// empty query
	// TODO: do I want to handle WILD and a tag? Actual regex?
	if query.Op == WILD && q == "" {
		// NOTE: this is all files with at least one tag and therefore of value:
		set.Union(slices.Collect(maps.Values(tagmap))...)
		return set
	}

	// NOTE: clone so that we don't accidentally overwrite the incoming tagmap
	set = maps.Clone(tagmap[q])
	// when queries < 2, this won't run
	for i := 1; i < len(query.Tags); i++ {
		q = query.Tags[i]
		switch query.Op {
		case OR:
			set.Union(tagmap[q])
		case AND:
			set.Intersect(tagmap[q])
		}
	}
	return set
}

// inverts the filelist using the full list from entries. works with intersected queries as long as
// processQueries is called first.
func invert(entries []Entry, files Set) Set {
	set := Set{}
	for _, e := range entries {
		if _, ok := files[e.filename]; !ok {
			set.Add(e.filename)
		}
	}
	return set
}

// adjacencies is a map from tag to a map of other tags occuring in the given files.
func makeAdjacencies(entries []Entry, files Set) map[string]map[string]Set {
	adjacencies := map[string]map[string]Set{}

	for _, e := range entries {
		// NOTE: this allows for a filelist shrunk after entries slice was made:
		if !files[e.filename] {
			continue
		}
		for i, tag := range e.tags {
			// make a slice copy but minus the current tag:
			others := make([]string, len(e.tags))
			copy(others, e.tags)
			others = slices.Delete(others, i, i+1)

			// allocate submap if necessary:
			if _, ok := adjacencies[tag]; !ok {
				adjacencies[tag] = map[string]Set{}
			}
			for _, other := range others {
				if _, ok := adjacencies[tag][other]; !ok {
					adjacencies[tag][other] = Set{}
				}
				adjacencies[tag][other].Add(e.filename)
			}
		}
	}
	return adjacencies
}

// reduces adjacencies to a single map[tag]Set not including the query tags
func reduceAdjacencies(adjacencies map[string]map[string]Set, query Query, invert bool) map[string]Set {
	reduced := map[string]Set{}
	if invert {
		// TODO: something's wrong here...
		// adjacencies keys will already reflect all tags built from an inverted filelist
		for _, adjmap := range adjacencies {
			for adj, files := range adjmap {
				if _, ok := reduced[adj]; !ok {
					reduced[adj] = Set{}
				}
				reduced[adj].Union(files)
			}
		}

		return reduced
	}
	for _, qtag := range query.Tags {
		for adjtag, files := range adjacencies[qtag] {
			if !slices.Contains(query.Tags, adjtag) {
				if _, ok := reduced[adjtag]; !ok {
					reduced[adjtag] = Set{}
				}
				reduced[adjtag].Add(files.Members()...)
			}
		}
	}
	return reduced
}

// prints out the intersected tagmap
func sprintFiles(files Set) string {
	ordered_files := make([]string, len(files))
	copy(ordered_files, slices.Collect(maps.Keys(files)))
	slices.Sort(ordered_files)
	// NOTE: I assume this is as efficient as strings.Builder :
	return fmt.Sprintln(strings.Join(ordered_files, "\n"))
}

func orderedTags(tagmap map[string]Set, query Query) []TagCount {
	ordered_tags := []TagCount{}
	// TODO: there's code smell about this whole approach.
	if query.Op == WILD {
		for q, s := range tagmap {
			ordered_tags = append(ordered_tags, TagCount{q, len(s)})
		}
	} else {
		for _, q := range query.Tags {
			ordered_tags = append(ordered_tags, TagCount{q, len(tagmap[q])})
		}
	}
	slices.SortFunc(ordered_tags, func(i, j TagCount) int {
		return cmp.Compare(i.count, j.count)
	})
	return ordered_tags
}

// prints out the complete and ordered collection of files, adjacencies, sums,
// and original query tags.
//
// format is a TOML syntax possibly useful elsewhere.
func printFiles(w io.Writer, entries []Entry, tagmap map[string]Set, files Set, adjacencies map[string]Set, query Query, verbose bool) {
	f := sprintFiles(files)
	if !verbose {
		fmt.Print(f)
		return
	}
	filesstr := fmt.Sprintln("[files]")
	filesstr += f

	tags := fmt.Sprintln("[tags]")
	otags := orderedTags(tagmap, query)
	tsb := strings.Builder{}
	// 20 * ' ' + '= 000' = 25
	tsb.Grow(len(otags) * 25)
	for _, t := range otags {
		tsb.WriteString(fmt.Sprintf("%-20s= %d\n", t.name, t.count))
	}
	tags += tsb.String()

	adj := fmt.Sprintln("[adjacencies]")
	oadj := orderedTags(adjacencies, Query{WILD, []string{}})
	asb := strings.Builder{}
	// 20 * ' ' + '= 000 : 000' = 31
	asb.Grow(len(oadj) * 31)
	for _, t := range oadj {
		// TODO: something's fucky about these len() with --invert :
		asb.WriteString(fmt.Sprintf("%-20s= %-3d : %d\n", t.name, t.count, len(tagmap[t.name])))
	}
	adj += asb.String()

	sums := fmt.Sprintln("[sums]")
	sums += fmt.Sprintf("files               = %-3d : %d\n", len(files), len(entries))
	sums += fmt.Sprintf("adjacencies         = %-3d : %d\n", len(adjacencies), len(tagmap))

	fmt.Fprintln(w, filesstr)
	fmt.Fprintln(w, tags)
	fmt.Fprintln(w, adj)
	fmt.Fprintln(w, sums)
}

func tag(opts options) {
	queries := parseQuery(opts.Query.Val)
	filelist := getFilelist(GLOB)
	entries := getEntries(filelist)

	// we shrink the entries list immediately if we want a date range:
	if opts.Date.Val != "" {
		entries = dateRange(entries, opts.Date.Val)
	}
	tagmap := makeTagmap(entries)

	// processQueries must precede invert because we want invert to respect combined tags:
	files := processQueries(tagmap, queries)
	if opts.Invert.Val {
		files = invert(entries, files)
	}
	// NOTE: the full makeAdjacencies map may one day be useful on its own
	adjacencies := reduceAdjacencies(makeAdjacencies(entries, files), queries, opts.Invert.Val)

	printFiles(os.Stdout, entries, tagmap, files, adjacencies, queries, opts.Verbose.Val)
}

func Tag(args []string) {
	opts := parseArgs(args)
	tag(opts)
}
