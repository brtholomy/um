package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	CMD     = "TEST"
	SUMMARY = "SUMMARY"
)

var helpErr = NewHelpError(CMD, SUMMARY)

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
	parseArgsInternal(helpErr, args, initOpts(), flags)
	assert.False(t, flags[0].IsSet())
	assert.False(t, flags[0].IsHelp())
	assert.True(t, flags[2].IsSet())
	assert.True(t, flags[4].IsHelp())

}

func TestParseArgsHelp(t *testing.T) {
	args := []string{"--help"}
	opts := initOpts()
	err := ParseArgs(helpErr, args, &opts)
	assert.ErrorContains(t, err, "um TEST [descriptor] [tags] [--source] [--write]\n\nSUMMARY")
	// we set the Val true although we don't expect to use it:
	assert.True(t, opts.Help.Val)
}

func TestExpandOpts(t *testing.T) {
	opts := initOpts()
	flags, err := expandOpts(&opts)
	assert.NoError(t, err)
	assert.Equal(t, &opts.Descriptor, flags[0])
	assert.Equal(t, &opts.Tags, flags[1])
	assert.Equal(t, &opts.Source, flags[2])
	assert.Equal(t, &opts.Write, flags[3])
	assert.Equal(t, &opts.Help, flags[4])
}

func TestExpandOptsNeedsPointer(t *testing.T) {
	opts := initOpts()
	_, err := expandOpts(opts)
	assert.ErrorContains(t, err, "needs a pointer to a struct")
	err = ParseArgs(helpErr, nil, opts)
	assert.ErrorContains(t, err, "needs a pointer to a struct")
}

func TestExpandOptsNeedsStruct(t *testing.T) {
	badOpts := struct {
		bar string
		foo bool
	}{"bar", false}

	_, err := expandOpts(&badOpts)
	assert.ErrorContains(t, err, "field must be a struct")
}

func TestParseArgsString(t *testing.T) {
	args := []string{"--source", "foo"}
	opts := initOpts()
	err := ParseArgs(helpErr, args, &opts)
	assert.NoError(t, err)
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
			err := ParseArgs(helpErr, tc.args, &opts)
			assert.NoError(t, err)
			if tc.boolVal != nil {
				assert.True(t, *tc.boolVal)
			}
			if tc.stringVal != nil {
				assert.Equal(t, *tc.stringVal, tc.stringWant)
			}
		})
	}
}
