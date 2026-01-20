package tag

import "strings"

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
