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
	"os"

	"github.com/spf13/cobra"
)

type bpdOptions struct {
	all         bool
	services    []string
	environment string
	repository  string
	key         string
	debug       bool
	parallel    int
	force       bool
}

func newBuildPushDeployCmd(out io.Writer) *cobra.Command {
	o := &bpdOptions{}

	cmd := &cobra.Command{
		Use:   "bpd",
		Short: "build, push & deploy!",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			buildArgs := os.Args[2:]

			buildCmd := newBuildCmd(out)
			buildCmd.ParseFlags(buildArgs)
			buildCmd.Run(cmd, buildArgs)

			pushCmd := newPushCmd(out)
			pushCmd.ParseFlags(buildArgs)
			pushCmd.Run(cmd, buildArgs)

			deployCmd := newDeployCmd(out)
			deployCmd.ParseFlags(buildArgs)
			deployCmd.Run(cmd, buildArgs)
		},
	}

	f := cmd.Flags()

	f.BoolVarP(&o.all, "all", "a", false, "build all services (default)")
	f.StringVarP(&o.environment, "environment", "e", "", "define build enviroment")
	f.StringVarP(&o.repository, "repository", "r", "", "repository to tag image with")
	f.StringVarP(&o.key, "key", "k", "latest", "key to tag latest image with")
	f.StringArrayVarP(&o.services, "service", "s", []string{}, "services to build")
	f.BoolVarP(&o.debug, "debug", "d", false, "debug output")
	f.IntVarP(&o.parallel, "parallel", "p", 4, "number of builds to run parallel")
	f.BoolVarP(&o.force, "force", "f", false, "force deploy")

	registerServiceAutocomplete(cmd)

	return cmd
}
