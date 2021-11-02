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
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gammazero/workerpool"
	"github.com/schollz/progressbar/v3"
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
	parallel    int
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

			if kipProject.Template() == "service" {
				o.services = []string{}
				o.all = true
			}

			if !o.all && len(o.services) == 0 {
				fmt.Fprint(out, "specify what to build using -s required or use -a to build all services\n")
				os.Exit(1)
			}

			if o.environment == "" {
				o.environment = kipProject.Environment()
			}

			if o.repository == "" {
				o.repository, _ = kipProject.Repository(o.environment)
			}

			if o.all && len(o.services) > 0 {
				fmt.Fprintf(out, "WARN: --all is ignored when --service is used\n")
				o.all = false
			}

			// add default docker build args
			extraArgs = append(kipProject.DockerBuildArgs(o.environment), extraArgs...)

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

			fmt.Fprintf(out, "Building services: %s\n", strings.Join(serviceNames, ","))

			preBuildscripts := kipProject.GetScripts("pre-build", o.environment)

			if len(preBuildscripts) > 0 {
				for _, script := range preBuildscripts {
					fmt.Fprintf(out, color.BlueString("RUN script: \"%s\"\n"), script.Name)

					err := script.Run(out, []string{})
					if err != nil {
						log.Fatalf("error running script \"%s\": %v", script.Name, err)
					}
				}
			}

			buildServices(out, servicesToBuild, o.repository, o.key, extraArgs, o.environment, o.parallel, o.debug)

			postBuildscripts := kipProject.GetScripts("post-build", o.environment)

			if len(postBuildscripts) > 0 {
				for _, script := range postBuildscripts {
					fmt.Fprintf(out, color.BlueString("RUN script: \"%s\"\n"), script.Name)

					err := script.Run(out, []string{})
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
	f.IntVarP(&o.parallel, "parallel", "p", 4, "number of builds to run parallel")

	registerServiceAutocomplete(cmd)

	return cmd
}

func removeStringFromArray(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func buildServices(out io.Writer, services []project.ServiceProject, repository string, key string, args []string, environment string, parallel int, debug bool) {
	wp := workerpool.New(parallel)

	os.Setenv("DOCKER_BUILDKIT", "1")

	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(out),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionUseANSICodes(true),
	)

	finished := 0
	building := []string{}
	total := 0
	start := time.Now()

	go func() {
		for !bar.IsFinished() {
			bar.RenderBlank()
			time.Sleep(time.Millisecond * 100)
		}
	}()

	for _, service := range services {
		service := service
		if service.HasDockerfile() {
			total++
			wp.Submit(func() {
				serviceStart := time.Now()
				building = append(building, service.Name())
				sort.Strings(building)

				bar.Describe(fmt.Sprintf("%v/%v Building (%v)", finished, total, strings.Join(building, ", ")))

				defer func() {
					finished++
					building = removeStringFromArray(building, service.Name())
					bar.Describe(fmt.Sprintf("%v/%v Building (%v)", finished, total, strings.Join(building, ", ")))
				}()

				extraArgs := append(args, service.DockerBuildArgs(environment)...)

				if debug {
					bar.Clear()
					fmt.Fprintf(out, color.BlueString("BUILDING %s with args: %s\n"), service.Name(), color.YellowString("%s", strings.Join(extraArgs, ", ")))
				}

				output, buildErr := service.Build(repository, key, extraArgs, environment)
				d := time.Since(serviceStart)
				d = d.Round(time.Millisecond)

				if buildErr == nil {
					bar.Clear()
					fmt.Fprintf(out, color.BlueString("BUILD %s %s %s\n"), service.Name(), color.GreenString("SUCCESS"), color.YellowString("%s", d))
					if debug {
						bar.Clear()
						fmt.Fprintf(out, "%v\n", string(output))
					}
				} else {
					bar.Clear()
					fmt.Fprintf(out, color.BlueString("BUILD %s %s %s\n"), service.Name(), color.RedString("FAILED"), color.YellowString("%s", d))
					bar.Clear()
					fmt.Fprintf(out, "%v\n", string(output))
				}
			})
		} else {
			bar.Clear()
			fmt.Fprintf(out, color.BlueString("SKIP service: \"%s\" no Dockerfile\n"), service.Name())
		}
	}

	wp.StopWait()
	d := time.Since(start)
	d = d.Round(time.Millisecond)
	bar.Finish()

	fmt.Fprintf(out, color.GreenString("BUILD %s\n"), color.YellowString("%s", d))
}
