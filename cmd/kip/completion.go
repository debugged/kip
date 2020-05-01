/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/spf13/cobra"
)

const defaultBoilerPlate = `
# Copyright 2020 The Kip Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`

var (
	completionLong = `
	Output shell completion code for the specified shell (bash or zsh).
	The shell code must be evaluated to provide interactive
	completion of kip commands.  This can be done by sourcing it from
	the .bash_profile.
	Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2`

	completionExample = `
		# Installing bash completion on macOS using homebrew
		## If running Bash 3.2 included with macOS
		    brew install bash-completion
		## or, if running Bash 4.1+
		    brew install bash-completion@2
		## If kip is installed via homebrew, this should start working immediately.
		## If you've installed via other means, you may need add the completion to your completion directory
		    kip completion bash > $(brew --prefix)/etc/bash_completion.d/kip
		# Installing bash completion on Linux
		## If bash-completion is not installed on Linux, please install the 'bash-completion' package
		## via your distribution's package manager.
		## Load the kip completion code for bash into the current shell
		    source <(kip completion bash)
		# Load the kip completion code for zsh[1] into the current shell
		    source <(kip completion zsh)
		# Set the kip completion code for zsh[1] to autoload on startup
		    kip completion zsh > "${fpath[1]}/_kip"`
)

var (
	completionShells = map[string]func(out io.Writer, cmd *cobra.Command) error{
		"bash": runCompletionBash,
		"zsh":  runCompletionZsh,
	}
)

// NewCmdCompletion creates the `completion` command
func newCompletionCmd(out io.Writer) *cobra.Command {
	shells := []string{}
	for s := range completionShells {
		shells = append(shells, s)
	}

	cmd := &cobra.Command{
		Use:                   "completion SHELL",
		DisableFlagsInUseLine: true,
		Short:                 "Output shell completion code for the specified shell (bash or zsh)",
		Long:                  completionLong,
		Example:               completionExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunCompletion(out, cmd, args)
			if err != nil {
				log.Fatalln(err)
			}
		},
		ValidArgs: shells,
	}

	return cmd
}

// RunCompletion checks given arguments and executes command
func RunCompletion(out io.Writer, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("Shell not specified.")
	}
	if len(args) > 1 {
		return errors.New("Too many arguments. Expected only the shell type.")
	}
	run, found := completionShells[args[0]]
	if !found {
		return errors.New(fmt.Sprintf("Unsupported shell type %q.", args[0]))
	}

	return run(out, cmd.Parent())
}

func runCompletionBash(out io.Writer, kip *cobra.Command) error {
	if _, err := out.Write([]byte(defaultBoilerPlate)); err != nil {
		return err
	}

	return kip.GenBashCompletion(out)
}

func runCompletionZsh(out io.Writer, kip *cobra.Command) error {
	zshHead := "#compdef kip\n"

	out.Write([]byte(zshHead))

	if _, err := out.Write([]byte(defaultBoilerPlate)); err != nil {
		return err
	}

	zshInitialization := `
__kip_bash_source() {
	alias shopt=':'
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__kip_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__kip_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__kip_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__kip_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__kip_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__kip_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__kip_filedir() {
	# Don't need to do anything here.
	# Otherwise we will get trailing space without "compopt -o nospace"
	true
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q 'GNU\|BusyBox'; then
	LWORD='\<'
	RWORD='\>'
fi
__kip_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__kip_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__kip_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__kip_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__kip_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__kip_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__kip_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	out.Write([]byte(zshInitialization))

	buf := new(bytes.Buffer)
	kip.GenBashCompletion(buf)
	out.Write(buf.Bytes())

	zshTail := `
BASH_COMPLETION_EOF
}
__kip_bash_source <(__kip_convert_bash_to_zsh)
`
	out.Write([]byte(zshTail))
	return nil
}
