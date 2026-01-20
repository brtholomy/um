package tag

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

// TODO: consider using this in tagmap and adjacencies. Currently only for printed results for the
// sake of slices.SortFunc
type TagCount struct {
	name  string
	count int
}

// just-in-time sort of our tag list for the sake of printFiles
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

// prints out the intersected tagmap
func sprintFiles(files Set) string {
	ordered_files := make([]string, len(files))
	copy(ordered_files, slices.Collect(maps.Keys(files)))
	slices.Sort(ordered_files)
	// NOTE: I assume this is as efficient as strings.Builder :
	return fmt.Sprintln(strings.Join(ordered_files, "\n"))
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
