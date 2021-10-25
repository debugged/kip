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
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var kipProject project.Project
var hasKipConfig = false

var rootPath, _ = filepath.Abs("/")
var rootCmd *cobra.Command

func main() {
	rootCmd = newRootCmd(color.Output, os.Args[1:])

	cobra.OnInitialize(initConfig)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	kipProject, err = loadKipProject(wd)

	hasKipConfig = err == nil
}

func loadKipProject(path string) (project.Project, error) {
	var kipProject project.Project
	var err error

	projectConfig := viper.New()
	projectConfig.AddConfigPath(path)
	projectConfig.SetConfigName("kip_config")
	projectConfig.SetConfigType("yaml")

	if err := projectConfig.ReadInConfig(); err != nil {
		newPath, err := filepath.Abs(filepath.Join(path, ".."))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if path == rootPath {
			return nil, errors.New("kip_config not found")
		}

		return loadKipProject(newPath)
	}

	env, _ := godotenv.Read()
	kipProject, err = project.GetProject(path, projectConfig, env)

	return kipProject, err
}
