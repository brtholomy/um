package cmd

type Subcommand string

const (
	Tag  Subcommand = "tag"
	Next Subcommand = "next"
	Last Subcommand = "last"
	Sort Subcommand = "sort"
	Help Subcommand = "help"
)
