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
	"debugged-dev/kip/v1/internal/version"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type initOptions struct {
	template string
}

func newInitCmd() *cobra.Command {
	o := &initOptions{}

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "A brief description of your command",
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
	
			wd = filepath.Join(wd, args[0])

			servicesDir := filepath.Join(wd, "services")
			
			os.MkdirAll(wd, os.ModePerm)
			os.MkdirAll(servicesDir, os.ModePerm)
	
			config := viper.New()
			config.AddConfigPath(wd)
			config.SetConfigName("kip_config")
			config.SetConfigType("yaml")
			
			config.Set("template", o.template)
			config.Set("version", version.Get().Version)
			
			config.SafeWriteConfig()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&o.template, "template", "t", "mono", "template for project mono or service")

	cmd.MarkFlagRequired("template")

	return cmd
}
