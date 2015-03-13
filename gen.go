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
	fmt.Fprintf(out, `#!/bin/bash


flags=()
two_word_flags=()
flags_with_completion=()
flags_completion=()
commands=()

__debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
        echo "$*" >> ${BASH_COMP_DEBUG_FILE}
    fi
}

__index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__handle_reply()
{
    __debug ${FUNCNAME}
    case $cur in
        -*)
            compopt -o nospace
            local allflags="${flags[*]} ${two_word_flags[*]}"
            COMPREPLY=( $(compgen -W "${allflags[*]}" -- "$cur") )
            [[ $COMPREPLY == *= ]] || compopt +o nospace
            return 0;
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    COMPREPLY=( $(compgen -W "${commands[*]}" -- "$cur") )
}

__handle_flags()
{
    if [[ $c -ge $cword ]]; then
        return
    fi
    __debug ${FUNCNAME} "c is" $c "words[c] is" ${words[c]}
    case ${words[c]} in
        -*)
            ;;
        *)
            return
            ;;
    esac

    # skip the argument to a two word flag
    if __contains_word "${words[c]}" "${two_word_flags[@]}"; then
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    # skip the flag itseld
    c=$((c+1))
    __handle_flags

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

func writeFlagHandler(name string, annotations map[string][]string, out *bytes.Buffer) {
	for key, value := range annotations {
		switch key {
		case "bash_comp_filename_ext":
			fmt.Fprintf(out, "    flags_with_completion+=(%q)\n", name)

			ext := strings.Join(value, "|")
			ext = "_filedir '@(" + ext + ")'"
			fmt.Fprintf(out, "    flags_completion+=(%q)\n", ext)
		}
	}
}

func writeShortFlag(flag *pflag.Flag, out *bytes.Buffer) {
	b := (flag.Value.Type() == "bool")
	name := flag.Shorthand
	format := "    "
	if !b {
		format += "two_word_"
	}
	format += "flags+=(\"-%s\")\n"
	fmt.Fprintf(out, format, name)
	writeFlagHandler("-"+name, flag.Annotations, out)
}

func writeFlag(flag *pflag.Flag, out *bytes.Buffer) {
	b := (flag.Value.Type() == "bool")
	name := flag.Name
	format := "    flags+=(\"--%s"
	if !b {
		format += "="
	}
	format += "\")\n"
	fmt.Fprintf(out, format, name)
	writeFlagHandler("--"+name, flag.Annotations, out)
}

func setFlags(cmd *cobra.Command, out *bytes.Buffer) {
	fmt.Fprintf(out, `    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

`)
	cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		writeFlag(flag, out)
		if len(flag.Shorthand) > 0 {
			writeShortFlag(flag, out)
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
	fmt.Fprintf(out, "    command_path=${command_path}_%s\n", cmd.Name())
	setCommands(cmd, out)
	setFlags(cmd, out)
	fmt.Fprintf(out, "    __handle_flags\n")
	fmt.Fprintf(out, "    __debug ${FUNCNAME} $c $cword\n")
	fmt.Fprintf(out, `    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
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
