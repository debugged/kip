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
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type helmArgsOptions struct {
	all         bool
	environment string
	repository  string
	key         string
}

func newHelmArgsCmd(out io.Writer) *cobra.Command {
	o := &helmArgsOptions{}

	cmd := &cobra.Command{
		Use:   "helmargs",
		Short: "return image helm args",
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

			extraArgs := cmd.Flags().Args()

			services := kipProject.Services()

			if o.environment == "" {
				o.environment = kipProject.Environment()
			}

			if o.repository == "" {
				o.repository = kipProject.Repository()
			}

			imageArgs := []string{}

			for _, service := range services {
				if service.HasDockerfile() {
					var tag string = "latest"

					if o.key != "" {
						tag = "temp-" + o.key
					}

					buildID, err := service.GetImageID(tag, o.repository)

					if err != nil {
						fmt.Fprintln(out, err)
						fmt.Fprintf(out, color.RedString("image: \"%s\" not found be sure to run kip build first\n"), service.Name())
						os.Exit(1)
					}

					serviceKey := strings.ReplaceAll(service.Name(), "-", "_")

					imageArgs = append(imageArgs, []string{"--set", "global.services." + serviceKey + ".name=" + service.Name()}...)
					imageArgs = append(imageArgs, []string{"--set", "global.services." + serviceKey + ".tag=" + buildID}...)
				}
			}

			extraArgs = append(extraArgs, imageArgs...)
			fmt.Print(extraArgs)
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&o.all, "all", "a", true, "deploy all charts")
	f.StringVarP(&o.environment, "environment", "e", "", "define build enviroment")
	f.StringVarP(&o.repository, "repository", "r", "", "repository to tag image with")
	f.StringVarP(&o.key, "key", "k", "latest", "key to tag latest image with")

	return cmd
}
