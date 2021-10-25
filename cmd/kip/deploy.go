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
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"robpike.io/filter"
)

type deployOptions struct {
	all         bool
	charts      []string
	services    []string
	force       bool
	environment string
	repository  string
	key         string
}

func newDeployCmd(out io.Writer) *cobra.Command {
	o := &deployOptions{}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploys project",
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

			charts := kipProject.Charts()
			services := kipProject.Services()
			chartsToDeploy := []project.Chart{}
			servicesToDeploy := []project.ServiceProject{}

			if !o.all && len(o.charts) == 0 && len(o.services) == 0 {
				fmt.Fprint(out, "specify what to deploy using -c or -s required or use --all | -a to deploy all charts and services\n")
				os.Exit(1)
			}

			if o.all && len(o.charts) > 0 {
				o.all = false
			}

			if o.environment == "" {
				o.environment = kipProject.Environment()
			}

			if o.repository == "" {
				o.repository, _ = kipProject.Repository(o.environment)
			}

			if o.all {
				chartsToDeploy = append(chartsToDeploy, charts...)
				servicesToDeploy = append(servicesToDeploy, services...)
			} else {
				if len(o.charts) > 0 {

					for _, chartName := range o.charts {
						var foundChart *project.Chart = nil
						for _, chart := range charts {
							if chart.Name() == chartName {
								foundChart = &chart
								break
							}
						}

						if foundChart != nil {
							chartsToDeploy = append(chartsToDeploy, *foundChart)
						} else {
							fmt.Fprintf(out, "chart \"%s\" does not exist in project\n", chartName)
							os.Exit(1)
						}
					}
				}

				if len(o.services) > 0 {
					for _, serviceName := range o.services {
						var foundService project.ServiceProject = project.ServiceProject{}
						for _, service := range services {
							if service.Name() == serviceName {
								foundService = service
								break
							}
						}

						if foundService != (project.ServiceProject{}) {
							servicesToDeploy = append(servicesToDeploy, foundService)
						} else {
							fmt.Fprintf(out, "service \"%s\" does not exist in project\n", serviceName)
							os.Exit(1)
						}
					}
				}
			}

			preDeployscripts := kipProject.GetScripts("pre-deploy", o.environment)

			if len(preDeployscripts) > 0 {
				for _, script := range preDeployscripts {
					fmt.Fprintf(out, color.BlueString("RUN script: \"%s\"\n"), script.Name)

					err := script.Run(out, []string{})
					if err != nil {
						log.Fatalf("error running script \"%s\": %v", script.Name, err)
					}
				}
			}

			chartNames := filter.Apply(chartsToDeploy, func(c project.Chart) string {
				return c.Name()
			}).([]string)

			serviceNames := filter.Apply(servicesToDeploy, func(s project.ServiceProject) string {
				return s.Name()
			}).([]string)

			imageArgs := []string{}

			for _, service := range services {
				if service.HasDockerfile() {
					var tag string = "latest"

					if o.key != "" {
						tag = "temp-" + o.key
					}

					repo, _ := service.Repository(o.environment)
					buildID, err := service.GetImageID(tag, repo)

					if err != nil {
						fmt.Fprintln(out, err)
						fmt.Fprintf(out, color.RedString("image: \"%s\" not found be sure to run kip build first\n"), service.Name())
						os.Exit(1)
					}

					serviceKey := strings.ReplaceAll(service.Name(), "-", "_")

					imageArgs = append(imageArgs, []string{"--set", "global.services." + serviceKey + ".name=" + service.Name()}...)
					imageArgs = append(imageArgs, []string{"--set", "global.services." + serviceKey + ".tag=" + buildID}...)
				} else {
					fmt.Fprintf(out, color.BlueString("SKIP service: %s no Dockerfile\n"), service.Name())
				}
			}

			extraArgs = append(extraArgs, imageArgs...)

			fmt.Fprintf(out, "Deploying charts  : %s\n", strings.Join(chartNames, ","))

			if kipProject.Template() == "project" {
				fmt.Fprintf(out, "Deploying services: %s\n\n", strings.Join(serviceNames, ","))
			}
			deployCharts(out, chartsToDeploy, o.environment, extraArgs, o.force)
			if kipProject.Template() == "project" {
				deployServices(out, servicesToDeploy, o.environment, extraArgs, o.force)
			}

			postDeployscripts := kipProject.GetScripts("post-deploy", o.environment)

			if len(postDeployscripts) > 0 {
				for _, script := range postDeployscripts {
					fmt.Fprintf(out, color.BlueString("RUN script: \"%s\"\n"), script.Name)

					err := script.Run(out, []string{})
					if err != nil {
						log.Fatalf("error running script %s: %v", script.Name, err)
					}
				}
			}
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&o.all, "all", "a", false, "deploy all charts")
	f.StringVarP(&o.environment, "environment", "e", "", "define build enviroment")
	f.StringVarP(&o.repository, "repository", "r", "", "repository to tag image with")
	f.StringVarP(&o.key, "key", "k", "latest", "key to tag latest image with")
	f.StringArrayVarP(&o.charts, "charts", "c", []string{}, "charts to deploy")
	f.StringArrayVarP(&o.services, "service", "s", []string{}, "services to deploy")
	f.BoolVarP(&o.force, "force", "f", false, "force deploy")

	return cmd
}

func deployCharts(out io.Writer, charts []project.Chart, environment string, args []string, force bool) {
	for _, chart := range charts {
		isChanged, err := chart.IsChanged(environment, args)
		if err != nil {
			fmt.Fprint(out, err)
			os.Exit(1)
		}

		fmt.Fprintf(out, color.BlueString("DEPLOY chart: %s: %s \n"), chart.Name(), color.YellowString(environment))

		if !isChanged && !force {
			fmt.Fprintf(out, color.BlueString("DEPLOY chart: %s %s\n\n"), chart.Name(), color.YellowString("no changes"))
		} else {
			buildErr := chart.Deploy(environment, args)
			if buildErr == nil {
				fmt.Fprintf(out, color.BlueString("DEPLOY chart: %s %s\n\n"), chart.Name(), color.GreenString("SUCCESS"))
			} else {
				fmt.Fprint(out, buildErr)
				os.Exit(1)
			}
		}
	}
}

func deployServices(out io.Writer, services []project.ServiceProject, environment string, args []string, force bool) {
	for _, service := range services {
		charts := service.Charts()

		if len(charts) > 0 {
			fmt.Fprintf(out, color.BlueString("DEPLOY service: %s\n"), service.Name())
			deployCharts(out, charts, environment, args, force)
			fmt.Fprintf(out, color.BlueString("DEPLOY service: %s %s\n\n"), service.Name(), color.GreenString("SUCCESS"))
		} else {
			fmt.Fprintf(out, color.BlueString("SKIP DEPLOY service: \"%s\" no charts\n"), service.Name())
		}
	}
}
