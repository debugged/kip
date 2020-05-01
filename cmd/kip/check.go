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
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type command struct {
	name string
	info string
}

var dependencies = []command{
	{
		name: "kubectl",
		info: "https://kubernetes.io",
	},
	{
		name: "helm",
		info: "https://helm.sh",
	},
	{
		name: "docker",
		info: "https://www.docker.com",
	},
}

func newCheckCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "checks if all dependencies are in available in $PATH",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			hasError := false
			data := [][]string{}

			for _, command := range dependencies {
				path, err := exec.LookPath(command.name)

				if err != nil {
					data = append(data, []string{color.RedString("FAIL"), command.name, fmt.Sprintf("%s not found. for more info: %s", command.name, command.info)})
					hasError = true
				} else {
					data = append(data, []string{color.GreenString("PASS"), command.name, path})
				}
			}

			table := tablewriter.NewWriter(color.Output)
			table.SetHeader([]string{emoji.Sprint(":chicken:"), "command", "message"})

			for _, v := range data {
				table.Append(v)
			}
			table.Render()

			if hasError {
				os.Exit(1)
			}
		},
	}

	return cmd
}
