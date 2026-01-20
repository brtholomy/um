package tag

import (
	"maps"
	"slices"
	"strings"
	"time"
)

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
