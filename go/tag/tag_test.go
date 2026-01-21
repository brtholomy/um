package tag

import (
	"bytes"
	"testing"
	"time"

	// TODO: switch to something lighter: https://github.com/alecthomas/assert
	"github.com/stretchr/testify/assert"
)

const TEST_PATTERN string = "./testdata/*.md"

func TestParseHeader(t *testing.T) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	header := parseHeader(&entries[0].content)
	expected := "# 01.foo.md\n: 2024.09.25\n+ bar\n+ foo"
	assert.Equal(t, expected, header)
}

func TestEntriesLen(t *testing.T) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	expected := 6
	if len(entries) != expected {
		t.Errorf("entries should be len == %v, got %v", expected, len(entries))
	}
}

func TestEntries(t *testing.T) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	d, _ := time.Parse("2006.01.02", "2024.09.25")
	expected := Entry{filename: "01.foo.md", date: d, content: "# 01.foo.md\n: 2024.09.25\n+ bar\n+ foo\n\nFoo bar.\n", tags: []string{"bar", "foo"}}
	assert.Equal(t, expected, entries[0])
}

func TestTagmap(t *testing.T) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	tagmap := makeTagmap(entries)
	expected := Set{"01.foo.md": true, "02.foo.md": true, "03.bar.md": true}
	assert.Equal(t, expected, tagmap["bar"])
}

func TestAdjacencies(t *testing.T) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	tagmap := makeTagmap(entries)
	queries := parseQuery("bar")
	fs := processQueries(tagmap, queries)
	adjacencies := makeAdjacencies(entries, fs)
	expected := map[string]Set{
		"foo":     Set{"01.foo.md": true},
		"science": Set{"02.foo.md": true, "03.bar.md": true},
	}
	assert.Equal(t, expected, adjacencies["bar"])
}

func TestPrint(t *testing.T) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	tagmap := makeTagmap(entries)
	query := parseQuery("bar")
	fs := processQueries(tagmap, query)
	adjacencies := reduceAdjacencies(makeAdjacencies(entries, fs), query, false)
	buf := bytes.Buffer{}
	printFiles(&buf, entries, tagmap, fs, adjacencies, query, true)
	expected := `[files]
01.foo.md
02.foo.md
03.bar.md

[tags]
bar                 = 3

[adjacencies]
foo                 = 1   : 1
science             = 2   : 3

[sums]
files               = 3   : 6
adjacencies         = 2   : 4

`
	assert.Equal(t, expected, buf.String())
}

func TestBadTag(t *testing.T) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	tagmap := makeTagmap(entries)

	_, ok := tagmap["qaz"]
	assert.False(t, ok)
}

// Since entriesGlobOrStdin() involves filesystem reads, we test the underlying logic.
func BenchmarkParseContent(b *testing.B) {
	e := entriesGlobOrStdin(TEST_PATTERN)[0]
	for b.Loop() {
		parseContent(e.filename, &e.content)
	}
}

func BenchmarkTagmap(b *testing.B) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	for b.Loop() {
		makeTagmap(entries)
	}
}

func BenchmarkAdjacencies(b *testing.B) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	tagmap := makeTagmap(entries)
	queries := parseQuery("foo")
	fs := processQueries(tagmap, queries)
	for b.Loop() {
		makeAdjacencies(entries, fs)
	}
}

func BenchmarkPrint(b *testing.B) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	tagmap := makeTagmap(entries)
	query := parseQuery("bar")
	fs := processQueries(tagmap, query)
	adjacencies := reduceAdjacencies(makeAdjacencies(entries, fs), query, false)
	buf := bytes.Buffer{}
	for b.Loop() {
		printFiles(&buf, entries, tagmap, fs, adjacencies, query, true)
	}
}

func FuzzParseContent(f *testing.F) {
	entries := entriesGlobOrStdin(TEST_PATTERN)
	for _, e := range entries {
		f.Add(e.filename, e.content)
	}
	f.Fuzz(func(t *testing.T, filename, content string) {
		e := parseContent(filename, &content)
		if e.date.IsZero() {
			t.Fatalf("failed to read date: %v", filename)
		}
	})
}
