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
	"io"

	"github.com/spf13/cobra"
)

const (
	bashCompletionFunc = `
__kip_get_script()
{
    local kip_out
    if kip_out=$(kip script list -s); then
        COMPREPLY+=( $( compgen -W "${kip_out[*]}" -- "$cur" ) )
    fi
}

__kip_custom_func() {
    case ${last_command} in
        kip_run)
            __kip_get_script
            return
            ;;
        *)
            ;;
    esac
}
`
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func newRootCmd(out io.Writer, args []string) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "kip",
		Short: "The Kubernetes project manager.",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		BashCompletionFunction: bashCompletionFunc,
	}

	cmd.AddCommand(
		newNewCmd(out),
		newServiceCmd(out),
		newChartCmd(out),
		newScriptCmd(out),
		newRunCmd(out),
		newBuildCmd(out),
		newDeployCmd(out),
		newPushCmd(out),
		newGeneratorsCmd(out),
		newCheckCmd(out),
		newVersionCmd(out),
		newCompletionCmd(out),
	)

	return cmd
}
