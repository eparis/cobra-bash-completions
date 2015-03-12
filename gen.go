package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func preamble(out *bytes.Buffer) {
	fmt.Fprintf(out, "#!/bin/bash\n\n")
	fmt.Fprintf(out, "flags=()\n")
	fmt.Fprintf(out, "commands=()\n\n")
	fmt.Fprintf(out,
		`__handle_reply()
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
	fmt.Fprintf(out,
		`__start()
{
    local cur prev words cword split
    _init_completion -s || return

    local completions_func
    local c=0

    _%s
}

`, name)
	fmt.Fprintf(out, "complete -F __start %s\n", name)
	fmt.Fprintf(out, "# ex: ts=4 sw=4 et filetype=sh\n")
}

func setCommands(cmd *cobra.Command, out *bytes.Buffer) {
	fmt.Fprintf(out, "    commands=()\n")
	for _, c := range cmd.Commands() {
		fmt.Fprintf(out, "    commands+=(%s)\n", c.Name())
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
	fmt.Fprintf(out, "_%s()\n{\n", cmd.Name())
	fmt.Fprintf(out, "    c=$((c+1))\n")
	setCommands(cmd, out)
	setFlags(cmd, out)
	fmt.Fprintf(out,
		`    if [[ $c -lt $cword ]]; then
        completions_func=_${words[c]}
        declare -F $completions_func >/dev/null && $completions_func
	return
    fi

`)
	fmt.Fprintf(out, "    __handle_reply\n")
	fmt.Fprintf(out, "}\n\n")
}

func GenCompletion(cmd *cobra.Command, out *bytes.Buffer) {
	preamble(out)
	gen(cmd, out)
	postscript(out, cmd.Name())
}

func main() {
	kubectl := cmd.NewFactory(nil).NewKubectlCommand(os.Stdin, ioutil.Discard, ioutil.Discard)

	out := new(bytes.Buffer)

	GenCompletion(kubectl, out)

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
