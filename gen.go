package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd"

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

    local completions_func command_path
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
		fmt.Fprintf(out, "    commands+=(%q)\n", c.Name())
	}
	fmt.Fprintf(out, "\n")
}

func writeFlag(name string, b bool, short bool, out *bytes.Buffer) {
	format := "    flags+=(\"-"
	if ! short {
		format += "-"
	}
	format += "%s"
	if ! b && ! short {
		format += "="
	}
	format += "\")\n"
	fmt.Fprintf(out, format, name)
}

func setFlags(cmd *cobra.Command, out *bytes.Buffer) {
	fmt.Fprintf(out, "    flags=()\n")
	cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		b := (flag.Value.Type() == "Boolean")
		writeFlag(flag.Name, b, false, out)
		if len(flag.Shorthand) > 0 {
			writeFlag(flag.Shorthand, b, true, out)
		}
	})

	fmt.Fprintf(out, "\n")
}

func gen(cmd *cobra.Command, out *bytes.Buffer) {
	for _, c := range cmd.Commands() {
		gen(c, out)
	}
	commandName := cmd.CommandPath()
	commandName = strings.Replace(commandName, " ", "_", -1)
	fmt.Fprintf(out, "_%s()\n{\n", commandName)
	fmt.Fprintf(out, "    c=$((c+1))\n")
	setCommands(cmd, out)
	setFlags(cmd, out)
	fmt.Fprintf(out, `    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        declare -F $command_path >/dev/null && $command_path
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
