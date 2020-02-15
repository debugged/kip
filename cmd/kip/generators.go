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
	"debugged-dev/kip/v1/internal/generator"
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newGeneratorsCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generators",
		Short: "lists all available generators for creating services",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			data := [][]string{}

			for _, gen := range generator.Generators {
				generatorName := gen.Name

				if gen.Name == generator.DefaultGenerator {
					generatorName = fmt.Sprintf("%s (default)", gen.Name)
				}

				data = append(data, []string{generatorName, gen.Info})
			}

			table := tablewriter.NewWriter(color.Output)
			table.SetHeader([]string{"generator", "info"})
			
			for _, v := range data {
					table.Append(v)
			}

			fmt.Fprintln(out, `Below a list with service generators, use those to easily add new services to your project!`)

			table.Render()
		},
	}
	
	return cmd
}
