package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd"
	//"github.com/openshift/origin/pkg/cmd/cli"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func preamble(out *bytes.Buffer) {
	fmt.Fprintf(out, `#!/bin/bash


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
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
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

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions=("${must_have_one_flag[@]}")
    elif [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions=("${must_have_one_noun[@]}")
    else
        completions=("${commands[@]}")
    fi
    COMPREPLY=( $(compgen -W "${completions[*]}" -- "$cur") )

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
        declare -F __custom_func >/dev/null && __custom_func
    fi
}

__handle_nouns()
{
    if [[ $c -ge $cword ]]; then
        return
    fi

    __debug ${FUNCNAME} "c is" $c "words[c] is" ${words[c]}

    if ! __contains_word "${words[c]}" "${commands[@]}"; then
        last_noun="${words[c]}"
    fi

    if __contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
        c=$((c+1))
        __handle_flags
    fi
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
            __handle_nouns
            return
            ;;
    esac

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __debug "looking for ${flagname}"
    if __contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

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
	fmt.Fprintf(out, "__start_%s()\n", name)
	fmt.Fprintf(out, `{
    local cur prev words cword split
    _init_completion -s || return

    local completions_func command_path
    local c=0
    local flags=()
    local two_word_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=()
    local must_have_one_flag=()
    local must_have_one_noun=()
    local last_command
    local last_noun

    _%s
}

`, name)
	fmt.Fprintf(out, "complete -F __start_%s %s\n", name, name)
	fmt.Fprintf(out, "# ex: ts=4 sw=4 et filetype=sh\n")
}

func writeCommands(cmd *cobra.Command, out *bytes.Buffer) {
	fmt.Fprintf(out, "    commands=()\n")
	for _, c := range cmd.Commands() {
		fmt.Fprintf(out, "    commands+=(%q)\n", c.Name())
	}
	fmt.Fprintf(out, "\n")
}

func writeFlagHandler(name string, annotations map[string][]string, out *bytes.Buffer) {
	for key, value := range annotations {
		switch key {
		case "BASH_COMP_filename_ext":
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

func writeFlags(cmd *cobra.Command, out *bytes.Buffer) {
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

func writeRequiredFlag(cmd *cobra.Command, out *bytes.Buffer) {
	fmt.Fprintf(out, "    must_have_one_flag=()\n")
	for key, value := range cmd.Annotations {
		switch key {
		case "BASH_COMP_one_required_flag":
			for _, flag := range value {
				fmt.Fprintf(out, "    must_have_one_flag+=(%q)\n", flag)
			}
		}
	}
}

func writeRequiredNoun(cmd *cobra.Command, out *bytes.Buffer) {
	fmt.Fprintf(out, "    must_have_one_noun=()\n")
	for key, value := range cmd.Annotations {
		switch key {
		case "BASH_COMP_one_required_noun":
			for _, noun := range value {
				fmt.Fprintf(out, "    must_have_one_noun+=(%q)\n", noun)
			}
		}
	}
}

func gen(cmd *cobra.Command, out *bytes.Buffer) {
	for _, c := range cmd.Commands() {
		gen(c, out)
	}
	commandName := cmd.CommandPath()
	commandName = strings.Replace(commandName, " ", "_", -1)
	fmt.Fprintf(out, "_%s()\n{\n", commandName)
	fmt.Fprintf(out, "    last_command=%q\n", commandName)
	fmt.Fprintf(out, "    c=$((c+1))\n")
	fmt.Fprintf(out, "    command_path=_%s\n", commandName)
	writeCommands(cmd, out)
	writeFlags(cmd, out)
	writeRequiredFlag(cmd, out)
	writeRequiredNoun(cmd, out)
	fmt.Fprintf(out, `    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
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
	if len(cmd.BashCompletionFunction) > 0 {
		fmt.Fprintf(out, "%s\n", cmd.BashCompletionFunction)
	}
	gen(cmd, out)
	postscript(out, cmd.Name())
}

func genTest(cmd *cobra.Command, filename string) {
	out := new(bytes.Buffer)

	GenCompletion(cmd, out)

	outFile, err := os.Create(filename)
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

func main() {
	kubectl := cmd.NewFactory(nil).NewKubectlCommand(os.Stdin, ioutil.Discard, ioutil.Discard)
	genTest(kubectl, "/tmp/kubectl")

	//osc := cli.NewCommandCLI("osc", "osc")
	//genTest(osc, "/tmp/osc")
}
