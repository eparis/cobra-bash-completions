package main

import (
	"github.com/spf13/cobra"
)

func cmd(name string) *cobra.Command {
	return &cobra.Command{
		Use: name,
	}
}

func main() {
	l1 := cmd("1")
	l2a := cmd("2a")
	l2b := cmd("2b")
	l2a1 := cmd("2a1")

	l2a.AddCommand(l2a1)
	l1.AddCommand(l2a)
	l1.AddCommand(l2b)

	return
}
