package main

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func preamble(out *bytes.Buffer) {
	fmt.Fprintf(out, "#!/bin/bash\n\n")
	fmt.Fprintf(out, "flags=()\n")
	fmt.Fprintf(out, "commands=()\n\n")
	fmt.Fprintf(out, `__handle_reply()
{
    case $cur in
        -*)
            compopt -o nospace
            COMPREPLY=( $(compgen -W "${flags[*]}" -- "$cur") )
            [[ $COMPREPLY == *= ]] || compopt +o nospace
            return 0;
            ;;
    esac

    COMPREPLY=( $(compgen -W "${commands[*]}" -- "$cur") )
}

`)
}

func postscript(out *bytes.Buffer, name string) {
	fmt.Fprintf(out, `__start()
{
    local cur prev words cword split
    _init_completion -s || return
    _%s
}

`, name)

	fmt.Fprintf(out, "complete -F __start %s\n", name)
	fmt.Fprintf(out, "# ex: ts=4 sw=4 et filetype=sh\n")
}

func setCommands(cmd *cobra.Command, out *bytes.Buffer) {
	fmt.Fprintf(out, "    commands=()\n")
	for _, c := range cmd.Commands() {
		fmt.Fprintf(out, "    commands+=(%s)\n", c.Use)
	}
	fmt.Fprintf(out, "\n")
}

func setFlags(cmd *cobra.Command, out *bytes.Buffer) {
	fmt.Fprintf(out, "    flags=()\n")
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		fmt.Fprintf(out, "    flags+=(%s)\n", flag.Name)
		if len(flag.Shorthand) > 0 {
			fmt.Fprintf(out, "    flags+=(%s)\n", flag.Shorthand)
		}
	})

	fmt.Fprintf(out, "\n")
}

func gen(cmd *cobra.Command, out *bytes.Buffer) {
	for _, c := range cmd.Commands() {
		gen(c, out)
	}
	fmt.Fprintf(out, "_%s()\n{\n", cmd.Use)
	setCommands(cmd, out)
	setFlags(cmd, out)
	fmt.Fprintf(out, "    __handle_reply\n")
	fmt.Fprintf(out, "}\n\n")
}

func GenCompletion(cmd *cobra.Command, out *bytes.Buffer) {
	preamble(out)
	gen(cmd, out)
	postscript(out, cmd.Use)
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

func main() {
	l1 := cmd(path.Base(os.Args[0]))
	l2a := cmd("2a")
	l2b := cmd("2b")
	l2a1 := cmd("2a1")

	l2a.AddCommand(l2a1)
	l1.AddCommand(l2a)
	l1.AddCommand(l2b)

	out := new(bytes.Buffer)

	GenCompletion(l1, out)

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
