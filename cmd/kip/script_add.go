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
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/spf13/cobra"
)

type addScriptOptions struct {
	service string
	command string
	bindings []string
}

func newAddScriptCmd(out io.Writer) *cobra.Command {
	o := &addScriptOptions{}

	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "adds a new script to your project or service",
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
			if !hasKipConfig {
				log.Fatalln("run this command inside a kip project")
			}

			scriptName := args[0]

			var err error

			if o.service != "" && kipProject.Template() == "project" {
				service, err := kipProject.GetService(o.service)

				if err != nil {
					log.Fatal(err)
				}

				err = service.AddScript(scriptName, o.command, o.bindings)

				if err != nil {
					log.Fatal(err)
				}
			}else {
				err = kipProject.AddScript(scriptName, o.command, o.bindings)

				if err != nil {
					log.Fatal(err)
				}
			}

			fmt.Printf("script config added")
		},
	}

	f := cmd.Flags()

	f.StringVarP(&o.service, "service", "s", "", "service where to add script")
	f.StringVarP(&o.command, "command", "c", "", "command to initiate script with, for example: bash, sh, node, python")
	f.StringArrayVarP(&o.bindings, "bind", "b", []string{}, "script bindings, for example: pre-build, post-build, pre-deploy, post-deploy")

	cmd.MarkFlagRequired("command")

	return cmd
}
