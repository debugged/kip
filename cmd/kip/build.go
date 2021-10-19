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
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/gammazero/workerpool"
	"github.com/spf13/cobra"
	"robpike.io/filter"
)

type buildOptions struct {
	all         bool
	services    []string
	environment string
	repository  string
	key         string
	debug       bool
}

func newBuildCmd(out io.Writer) *cobra.Command {
	o := &buildOptions{}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "builds a service",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !hasKipConfig {
				log.Fatalln("run this command inside a kip project")
			}

			extraArgs := cmd.Flags().Args()

			services := kipProject.Services()
			servicesToBuild := []project.ServiceProject{}

			if !o.all && len(o.services) == 0 {
				o.all = true
			}

			if o.environment == "" {
				o.environment = kipProject.Environment()
			}

			if o.repository == "" {
				o.repository = kipProject.Repository(o.environment)
			}

			if o.all && len(o.services) > 0 {
				fmt.Fprintf(out, "WARN: --all is ignored when --service is used\n")
				o.all = false
			}

			if o.all {
				servicesToBuild = append(servicesToBuild, services...)
			} else if len(o.services) > 0 {

				for _, serviceName := range o.services {
					var foundService project.Project = nil
					for _, service := range services {
						if service.Name() == serviceName {
							foundService = service
							break
						}
					}

					if foundService != nil {
						servicesToBuild = append(servicesToBuild, foundService.(project.ServiceProject))
					} else {
						fmt.Fprintf(out, "service \"%s\" does not exist in project\n", serviceName)
						os.Exit(1)
					}
				}
			}

			serviceNames := filter.Apply(servicesToBuild, func(s project.ServiceProject) string {
				return s.Name()
			}).([]string)

			fmt.Fprintf(out, "Building services: %s\n\n", strings.Join(serviceNames, ","))

			preBuildscripts := kipProject.GetScripts("pre-build", o.environment)

			if len(preBuildscripts) > 0 {
				for _, script := range preBuildscripts {
					fmt.Fprintf(out, color.BlueString("RUN script: \"%s\"\n"), script.Name)

					err := script.Run([]string{})
					if err != nil {
						log.Fatalf("error running script \"%s\": %v", script.Name, err)
					}
				}
			}

			buildServices(out, servicesToBuild, o.repository, o.key, extraArgs, o.environment, o.debug)

			postBuildscripts := kipProject.GetScripts("post-build", o.environment)

			if len(postBuildscripts) > 0 {
				for _, script := range postBuildscripts {
					fmt.Fprintf(out, color.BlueString("RUN script: \"%s\"\n"), script.Name)

					err := script.Run([]string{})
					if err != nil {
						log.Fatalf("error running script \"%s\": %v", script.Name, err)
					}
				}
			}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&o.all, "all", "a", false, "build all services (default)")
	f.StringVarP(&o.environment, "environment", "e", "", "define build enviroment")
	f.StringVarP(&o.repository, "repository", "r", "", "repository to tag image with")
	f.StringVarP(&o.key, "key", "k", "latest", "key to tag latest image with")
	f.StringArrayVarP(&o.services, "service", "s", []string{}, "services to build")
	f.BoolVarP(&o.debug, "debug", "d", false, "debug output")

	return cmd
}

func buildServices(out io.Writer, services []project.ServiceProject, repository string, key string, args []string, environment string, debug bool) {
	wp := workerpool.New(runtime.NumCPU())

	os.Setenv("DOCKER_BUILDKIT", "1")

	for _, service := range services {
		service := service
		if service.HasDockerfile() {
			wp.Submit(func() {
				fmt.Fprintf(out, color.BlueString("BUILD service: \"%s\"\n"), service.Name())
				output, buildErr := service.Build(repository, key, args, environment, debug)
				if buildErr == nil {
					fmt.Fprintf(out, color.BlueString("BUILD %s %s\n"), service.Name(), color.GreenString("SUCCESS"))
				} else {
					fmt.Fprintf(out, color.BlueString("BUILD %s %s\n"), service.Name(), color.RedString("FAILED"))
					fmt.Fprint(out, string(output))
					wp.Stop()
				}
			})
		} else {
			fmt.Fprintf(out, color.BlueString("SKIP service: \"%s\" no Dockerfile\n"), service.Name())
		}
	}

	wp.StopWait()
}
