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
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"robpike.io/filter"
)

type addRunOptions struct {
	service string
}

func newRunCmd(out io.Writer) *cobra.Command {
	o := &addRunOptions{}

	var scriptNames []string
	if kipProject != nil {
		scripts := kipProject.GetScripts("", "")
		scriptNames = filter.Apply(scripts, func(s project.Script) string {
			return s.Name
		}).([]string)
	}

	cmd := &cobra.Command{
		Use:   "run [script-name]",
		Short: "creates a new kubernetes project",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a script name argument")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if !hasKipConfig {
				fmt.Fprintln(out, color.RedString("run this command inside a kip project"))
				os.Exit(1)
			}

			scriptName := args[0]

			extraArgs := []string{}

			f := cmd.Flags()

			if f.ArgsLenAtDash() != -1 {
				extraArgs = f.Args()[f.ArgsLenAtDash():]
			}

			var err error
			project := kipProject

			if o.service != "" && kipProject.Template() == "project" {
				project, err = kipProject.GetService(o.service)

				if err != nil {
					log.Fatal(err)
				}
			}

			script, err := project.GetScript(scriptName)

			if err != nil {
				log.Fatal(err)
			}

			err = script.Run(extraArgs)

			if err != nil {
				log.Fatal(err)
			}
		},
		ValidArgs: scriptNames,
	}

	f := cmd.Flags()

	f.StringVarP(&o.service, "service", "s", "", "service of script")

	return cmd
}
