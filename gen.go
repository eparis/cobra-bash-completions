package main

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

func preamble(out *bytes.Buffer) {
	fmt.Fprintf(out, "#!/bin/bash\n")
}

func postscript(out *bytes.Buffer, name string) {
	fmt.Fprintf(out, "complete -F _%s %s\n", name, name)
	fmt.Fprintf(out, "# ex: ts=4 sw=4 et filetype=sh\n")
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func cmd(name string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: "short " + name,
		Long:  "long " + name,
		Run:   runHelp,
	}
}

func gen(cmd *cobra.Command, out *bytes.Buffer) {
	for _, c := range cmd.Commands() {
		gen(c, out)
	}
	fmt.Fprintf(out, "_%s()\n{\n", cmd.Use)
	fmt.Fprintf(out, "}\n\n")
}

func GenCompletion(cmd *cobra.Command, out *bytes.Buffer) {
	preamble(out)
	gen(cmd, out)
	postscript(out, cmd.Use)
}

func main() {
	l1 := cmd(path.Base(os.Args[0]))
	l2a := cmd("2a")
	l2b := cmd("2b")
	l2a1 := cmd("2a1")

	l2a.AddCommand(l2a1)
	l1.AddCommand(l2a)
	l1.AddCommand(l2b)

	out := new(bytes.Buffer)

	gen(l1, out)

	outFile, err := os.Create("/tmp/out.sh")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()
	_, err = outFile.Write(out.Bytes())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
