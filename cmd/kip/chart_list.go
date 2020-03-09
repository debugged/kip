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

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newListChartCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists all charts",
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

			switch kipProject.Template() {
			case "project":
				fmt.Printf("Charts in project %s\n", kipProject.Name())
				renderChartsTable(kipProject.Charts())

				for _, service := range kipProject.Services() {
					fmt.Printf("\nCharts in service: %s\n", service.Name())
					renderChartsTable(service.Charts())
				}

				break;
			case "service":
				fmt.Printf("Charts in service: %s\n", kipProject.Name())
				renderChartsTable(kipProject.Charts())
				break;
			}
		},
	}

	return cmd
}

func renderChartsTable(charts []project.Chart) {
	data := [][]string{}

	for _, chart := range charts {
		data = append(data, []string{chart.Name(), chart.Path()})
	}

	table := tablewriter.NewWriter(color.Output)
	table.SetHeader([]string{"charts", "info"})

	for _, v := range data {
		table.Append(v)
	}

	table.Render()
}
