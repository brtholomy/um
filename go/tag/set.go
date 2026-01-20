package tag

import (
	"maps"
	"slices"
)

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
