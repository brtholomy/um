package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	CMD     = "TEST"
	SUMMARY = "SUMMARY"
)

type options struct {
	Descriptor Arg
	Tags       Arg
	Source     String
	Write      Bool
	Help       Bool
}

func initOpts() options {
	return options{
		Arg{"", "midfix file descriptor"},
		Arg{"", "tags to add to new file"},
		String{"--source", "-s", "", "path to source list. reads from stdin if omitted."},
		Bool{"--write", "-w", false, "write sorted list back to --key file"},
		Bool{"--help", "-h", false, "show help"},
	}
}

func TestParseArgsInternal(t *testing.T) {
	flags := []Flag{
		&Arg{"", "midfix file descriptor"},
		&Arg{"", "tags to add to new file"},
		&String{"--source", "-s", "", "path to source list. reads from stdin if omitted."},
		&Bool{"--write", "-w", false, "write sorted list back to --key file"},
		&Bool{"--help", "-h", false, "show help"},
	}
	args := []string{"--source", "foo"}
	parseArgsInternal(CMD, SUMMARY, args, initOpts(), flags)
	assert.False(t, flags[0].IsSet())
	assert.False(t, flags[0].IsHelp())
	assert.True(t, flags[2].IsSet())
	assert.True(t, flags[4].IsHelp())

}

// TODO: return errors from Help routines and print at top level, so I can test those cases
func TestParseArgsString(t *testing.T) {
	args := []string{"--source", "foo"}
	opts := initOpts()
	ParseArgs(CMD, SUMMARY, args, &opts)
	assert.Equal(t, opts.Source.Val, "foo")
}

func TestParseArgs(t *testing.T) {
	opts := initOpts()
	cases := []struct {
		name       string
		args       []string
		stringVal  *string
		stringWant string
		boolVal    *bool
		boolWant   bool
	}{
		{"arg", []string{"bar"}, &opts.Descriptor.Val, "bar", nil, false},
		{"arg arg", []string{"bar", "foo"}, &opts.Tags.Val, "foo", nil, false},
		{"--bool", []string{"--write"}, nil, "", &opts.Write.Val, true},
		{"arg --bool", []string{"foo", "--write"}, &opts.Descriptor.Val, "foo", &opts.Write.Val, true},
		{"--flag val", []string{"--source", "foo"}, &opts.Source.Val, "foo", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ParseArgs(CMD, SUMMARY, tc.args, &opts)
			if tc.boolVal != nil {
				assert.True(t, *tc.boolVal)
			}
			if tc.stringVal != nil {
				assert.Equal(t, *tc.stringVal, tc.stringWant)
			}
		})
	}
}
