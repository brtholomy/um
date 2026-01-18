package sort

import (
	"flag"
	"fmt"

	"github.com/brtholomy/um/go/flags"
)

const (
	key flags.Flag = "key"
)

func Sort(args []string) {
	sortFlags := flag.NewFlagSet("sort", flag.ExitOnError)
	sortKey := sortFlags.String(string(key), "", "filename of ordered list to use as key")

	sortFlags.Parse(args)
	fmt.Println("subcmd sort")
	fmt.Printf("key: %#v\n", *sortKey)
	fmt.Printf("args: %#v\n", sortFlags.Args())
	sortFlags.Usage()
}
