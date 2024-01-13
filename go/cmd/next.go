/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
)

// nextCmd represents the next command
var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("next called")
		// shcmd := exec.Command("", "/home/bth/dev/um/bash/next/next.sh")

		// create a new *Cmd instance
		// command as the first arg, arguments as remaining args
		lscmd := exec.Command("ls")

		// The `Output` method executes the command and
		// collects the output, returning its value
		out, err := lscmd.Output()
		if err != nil {
			// if there was any error, print it here
			fmt.Println("could not run command: ", err)
		}
		// otherwise, print the output from running the command
		fmt.Println("Output: ", string(out))
	},
}

func init() {
	rootCmd.AddCommand(nextCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nextCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nextCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
