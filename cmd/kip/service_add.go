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
	"os"

	"debugged-dev/kip/v1/internal/project"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type addServiceOptions struct {
	generator string
}

func newAddServiceCmd(out io.Writer) *cobra.Command {
	o := &addServiceOptions{}

	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "generates a new service in your kip project",
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
				fmt.Fprintln(out, color.RedString("run this command inside a kip project"))
				os.Exit(1)
			}
			
			serviceName := args[0]

			f := cmd.Flags()
			
			extraArgs := []string{}

			if f.ArgsLenAtDash() != -1 {
				extraArgs = f.Args()[f.ArgsLenAtDash():]
			}

			if kipProject.Template() == "service" {
				fmt.Println("cannot create service inside kip service project")
				os.Exit(1)
			}

			serviceDir := kipProject.Paths().Services

			err := project.CreateServiceProject(serviceDir, serviceName, o.generator, extraArgs)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("service: %s created!\n", serviceName)
		},
	}

	f := cmd.Flags()

	f.StringVarP(&o.generator, "generator", "g", "", "generator for service")

	return cmd
}
