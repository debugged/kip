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
	"debugged-dev/kip/v1/internal/project"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"robpike.io/filter"
)

type listScriptOptions struct {
	simple bool
}

func newListScriptCmd(out io.Writer) *cobra.Command {
	o := &listScriptOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists all scripts",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !hasKipConfig {
				fmt.Fprintln(out, color.RedString("run this command inside a kip project"))
				os.Exit(1)
			}

			if o.simple {
				scripts := kipProject.GetScripts("", "")
				scriptNames := filter.Apply(scripts, func(s project.Script) string {
					return s.Name
				}).([]string)
				fmt.Println(strings.Join(scriptNames[:], " "))
				os.Exit(0)
			}

			switch kipProject.Template() {
			case "project":
				fmt.Printf("Scripts in project %s\n", kipProject.Name())
				renderScriptsTable(kipProject.GetScripts("", ""))

				for _, service := range kipProject.Services() {
					fmt.Printf("\nScripts in service: %s\n", service.Name())
					renderScriptsTable(service.GetScripts("", ""))
				}

				break
			case "service":
				fmt.Printf("Scripts in service: %s\n", kipProject.Name())
				renderScriptsTable(kipProject.GetScripts("", ""))
				break
			}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&o.simple, "simple", "s", false, "print simple output")
	f.MarkHidden("simple")

	return cmd
}

func renderScriptsTable(scripts []project.Script) {
	data := [][]string{}

	for _, script := range scripts {
		data = append(data, []string{script.Name, script.Command, strings.Join(script.Bindings, ","), script.Path})
	}

	table := tablewriter.NewWriter(color.Output)
	table.SetHeader([]string{"name", "command", "binding", "path"})

	for _, v := range data {
		table.Append(v)
	}

	table.Render()
}
