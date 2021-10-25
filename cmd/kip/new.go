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
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type newOptions struct {
	template  string
	generator string
}

func newNewCmd(out io.Writer) *cobra.Command {
	o := &newOptions{}

	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "creates a new kubernetes project",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a name argument")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			wd, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			projectName := args[0]

			switch o.template {
			case "project":
				err = project.CreateMonoProject(wd, projectName)
				break
			case "service":
				extraArgs := []string{}

				f := cmd.Flags()

				if f.ArgsLenAtDash() != -1 {
					extraArgs = f.Args()[f.ArgsLenAtDash():]
				}
				err = project.CreateServiceProject(wd, projectName, o.generator, extraArgs)
				break
			default:
				os.Exit(1)
			}

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("project: %s created!\n", projectName)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&o.template, "template", "t", "project", "project | service. A project can contain multiple services")
	f.StringVarP(&o.generator, "generator", "g", "", "generator used for creating service project")

	return cmd
}
