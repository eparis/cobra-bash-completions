#!/bin/bash


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
        flagname=${flagname%!=(MISSING)*} # strip everything after the =
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

# call kubectl get $1,
# use the first column in compgen
# we could use templates, but then would need a template per resource
__kubectl_parse_get()
{
    local kubectl_output out
    if kubectl_output=$(kubectl get --no-headers "$1" 2>/dev/null); then
        out=($(echo "${kubectl_output}" | awk '{print $1}'))
        COMPREPLY=( $( compgen -W "${out[*]}" -- "$cur" ) )
    fi
}

__kubectl_get_resource()
{
    __kubectl_parse_get ${last_noun}
    if [[ $? -eq 0 ]]; then
        return 0
    fi
}

__custom_func() {
    case ${last_command} in
        kubectl_get | kubectl_describe)
	    __kubectl_get_resource
            return 0
            ;;
        *)
            ;;
    esac
}

_kubectl_version()
{
    last_command="kubectl_version"
    c=$((c+1))
    command_path=_kubectl_version
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--client")
    flags+=("-c")
    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_clusterinfo()
{
    last_command="kubectl_clusterinfo"
    c=$((c+1))
    command_path=_kubectl_clusterinfo
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_proxy()
{
    last_command="kubectl_proxy"
    c=$((c+1))
    command_path=_kubectl_proxy
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--api-prefix=")
    flags+=("--help")
    flags+=("-h")
    flags+=("--port=")
    two_word_flags+=("-p")
    flags+=("--www=")
    two_word_flags+=("-w")
    flags+=("--www-prefix=")
    two_word_flags+=("-P")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_get()
{
    last_command="kubectl_get"
    c=$((c+1))
    command_path=_kubectl_get
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    flags+=("--no-headers")
    flags+=("--output=")
    two_word_flags+=("-o")
    flags+=("--output-version=")
    flags+=("--selector=")
    two_word_flags+=("-l")
    flags+=("--template=")
    two_word_flags+=("-t")
    flags+=("--watch")
    flags+=("-w")
    flags+=("--watch-only")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("secret")
    must_have_one_noun+=("replicationcontroller")
    must_have_one_noun+=("endpoints")
    must_have_one_noun+=("event")
    must_have_one_noun+=("namespace")
    must_have_one_noun+=("pod")
    must_have_one_noun+=("service")
    must_have_one_noun+=("status")
    must_have_one_noun+=("limitrange")
    must_have_one_noun+=("resourcequota")
    must_have_one_noun+=("node")
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_describe()
{
    last_command="kubectl_describe"
    c=$((c+1))
    command_path=_kubectl_describe
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("resourcequota")
    must_have_one_noun+=("pod")
    must_have_one_noun+=("replicationcontroller")
    must_have_one_noun+=("service")
    must_have_one_noun+=("minion")
    must_have_one_noun+=("node")
    must_have_one_noun+=("limitrange")
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_create()
{
    last_command="kubectl_create"
    c=$((c+1))
    command_path=_kubectl_create
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--filename=")
    flags_with_completion+=("--filename")
    flags_completion+=("_filedir '@(json|yaml|yml)'")
    two_word_flags+=("-f")
    flags_with_completion+=("-f")
    flags_completion+=("_filedir '@(json|yaml|yml)'")
    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_flag+=("-f")
    must_have_one_flag+=("--filename=")
    must_have_one_flag+=("-")
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_update()
{
    last_command="kubectl_update"
    c=$((c+1))
    command_path=_kubectl_update
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--filename=")
    two_word_flags+=("-f")
    flags+=("--help")
    flags+=("-h")
    flags+=("--patch=")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_delete()
{
    last_command="kubectl_delete"
    c=$((c+1))
    command_path=_kubectl_delete
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("--filename=")
    two_word_flags+=("-f")
    flags+=("--help")
    flags+=("-h")
    flags+=("--selector=")
    two_word_flags+=("-l")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_config_view()
{
    last_command="kubectl_config_view"
    c=$((c+1))
    command_path=_kubectl_config_view
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    flags+=("--merge")
    flags+=("--no-headers")
    flags+=("--output=")
    two_word_flags+=("-o")
    flags+=("--output-version=")
    flags+=("--template=")
    two_word_flags+=("-t")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_config_set-cluster()
{
    last_command="kubectl_config_set-cluster"
    c=$((c+1))
    command_path=_kubectl_config_set-cluster
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--embed-certs")
    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_config_set-credentials()
{
    last_command="kubectl_config_set-credentials"
    c=$((c+1))
    command_path=_kubectl_config_set-credentials
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--embed-certs")
    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_config_set-context()
{
    last_command="kubectl_config_set-context"
    c=$((c+1))
    command_path=_kubectl_config_set-context
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_config_set()
{
    last_command="kubectl_config_set"
    c=$((c+1))
    command_path=_kubectl_config_set
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_config_unset()
{
    last_command="kubectl_config_unset"
    c=$((c+1))
    command_path=_kubectl_config_unset
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_config_use-context()
{
    last_command="kubectl_config_use-context"
    c=$((c+1))
    command_path=_kubectl_config_use-context
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_config()
{
    last_command="kubectl_config"
    c=$((c+1))
    command_path=_kubectl_config
    commands=()
    commands+=("view")
    commands+=("set-cluster")
    commands+=("set-credentials")
    commands+=("set-context")
    commands+=("set")
    commands+=("unset")
    commands+=("use-context")

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--envvar")
    flags+=("--global")
    flags+=("--help")
    flags+=("-h")
    flags+=("--local")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_namespace()
{
    last_command="kubectl_namespace"
    c=$((c+1))
    command_path=_kubectl_namespace
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_log()
{
    last_command="kubectl_log"
    c=$((c+1))
    command_path=_kubectl_log
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--follow")
    flags+=("-f")
    flags+=("--help")
    flags+=("-h")
    flags+=("--interactive")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_rollingupdate()
{
    last_command="kubectl_rollingupdate"
    c=$((c+1))
    command_path=_kubectl_rollingupdate
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--filename=")
    two_word_flags+=("-f")
    flags+=("--help")
    flags+=("-h")
    flags+=("--poll-interval=")
    flags+=("--timeout=")
    flags+=("--update-period=")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_resize()
{
    last_command="kubectl_resize"
    c=$((c+1))
    command_path=_kubectl_resize
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--current-replicas=")
    flags+=("--help")
    flags+=("-h")
    flags+=("--replicas=")
    flags+=("--resource-version=")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_exec()
{
    last_command="kubectl_exec"
    c=$((c+1))
    command_path=_kubectl_exec
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--container=")
    two_word_flags+=("-c")
    flags+=("--help")
    flags+=("-h")
    flags+=("--pod=")
    two_word_flags+=("-p")
    flags+=("--stdin")
    flags+=("-i")
    flags+=("--tty")
    flags+=("-t")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_port-forward()
{
    last_command="kubectl_port-forward"
    c=$((c+1))
    command_path=_kubectl_port-forward
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    flags+=("--pod=")
    two_word_flags+=("-p")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_run-container()
{
    last_command="kubectl_run-container"
    c=$((c+1))
    command_path=_kubectl_run-container
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--dry-run")
    flags+=("--generator=")
    flags+=("--help")
    flags+=("-h")
    flags+=("--image=")
    flags+=("--labels=")
    two_word_flags+=("-l")
    flags+=("--no-headers")
    flags+=("--output=")
    two_word_flags+=("-o")
    flags+=("--output-version=")
    flags+=("--overrides=")
    flags+=("--port=")
    flags+=("--replicas=")
    two_word_flags+=("-r")
    flags+=("--template=")
    two_word_flags+=("-t")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_stop()
{
    last_command="kubectl_stop"
    c=$((c+1))
    command_path=_kubectl_stop
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--all")
    flags+=("--filename=")
    two_word_flags+=("-f")
    flags+=("--help")
    flags+=("-h")
    flags+=("--selector=")
    two_word_flags+=("-l")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_expose()
{
    last_command="kubectl_expose"
    c=$((c+1))
    command_path=_kubectl_expose
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--container-port=")
    flags+=("--create-external-load-balancer")
    flags+=("--dry-run")
    flags+=("--generator=")
    flags+=("--help")
    flags+=("-h")
    flags+=("--no-headers")
    flags+=("--output=")
    two_word_flags+=("-o")
    flags+=("--output-version=")
    flags+=("--overrides=")
    flags+=("--port=")
    flags+=("--protocol=")
    flags+=("--public-ip=")
    flags+=("--selector=")
    flags+=("--service-name=")
    flags+=("--template=")
    two_word_flags+=("-t")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl_label()
{
    last_command="kubectl_label"
    c=$((c+1))
    command_path=_kubectl_label
    commands=()

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    flags+=("--no-headers")
    flags+=("--output=")
    two_word_flags+=("-o")
    flags+=("--output-version=")
    flags+=("--overwrite")
    flags+=("--resource-version=")
    flags+=("--template=")
    two_word_flags+=("-t")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

_kubectl()
{
    last_command="kubectl"
    c=$((c+1))
    command_path=_kubectl
    commands=()
    commands+=("version")
    commands+=("clusterinfo")
    commands+=("proxy")
    commands+=("get")
    commands+=("describe")
    commands+=("create")
    commands+=("update")
    commands+=("delete")
    commands+=("config")
    commands+=("namespace")
    commands+=("log")
    commands+=("rollingupdate")
    commands+=("resize")
    commands+=("exec")
    commands+=("port-forward")
    commands+=("run-container")
    commands+=("stop")
    commands+=("expose")
    commands+=("label")

    flags=()
    two_word_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--alsologtostderr")
    flags+=("--api-version=")
    flags+=("--auth-path=")
    two_word_flags+=("-a")
    flags+=("--certificate-authority=")
    flags+=("--client-certificate=")
    flags+=("--client-key=")
    flags+=("--cluster=")
    flags+=("--context=")
    flags+=("--help")
    flags+=("-h")
    flags+=("--insecure-skip-tls-verify")
    flags+=("--kubeconfig=")
    flags+=("--log_backtrace_at=")
    flags+=("--log_dir=")
    flags+=("--log_flush_frequency=")
    flags+=("--logtostderr")
    flags+=("--match-server-version")
    flags+=("--namespace=")
    flags+=("--password=")
    flags+=("--server=")
    two_word_flags+=("-s")
    flags+=("--stderrthreshold=")
    flags+=("--token=")
    flags+=("--user=")
    flags+=("--username=")
    flags+=("--v=")
    flags+=("--validate")
    flags+=("--vmodule=")

    must_have_one_flag=()
    must_have_one_noun=()
    __handle_flags
    __debug ${FUNCNAME} $c $cword
    if [[ $c -lt $cword ]]; then
        command_path="${command_path}_${words[c]}"
        __debug "looking for " ${command_path}
        declare -F $command_path >/dev/null && $command_path && return
    fi

    __handle_reply
}

__start_kubectl()
{
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

    _kubectl
}

complete -F __start_kubectl kubectl
# ex: ts=4 sw=4 et filetype=sh
