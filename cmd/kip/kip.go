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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// todo: create struct with kip root path etc
var kipRoot, hasKipConfig = "", false

var rootPath, _ = filepath.Abs("/")

func main() {
	cmd := newRootCmd(os.Stdout, os.Args[1:])

	cobra.OnInitialize(initConfig)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	// Find home directory.
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	kipRoot, hasKipConfig = loadKipRoot(wd)
}

func loadKipRoot(path string) (response string, ok bool) {
	viper.AddConfigPath(path)
	viper.SetConfigName("kip_config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		newPath, err := filepath.Abs(filepath.Join(path, ".."))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if(path == rootPath) {
			return "", false
		}

		return loadKipRoot(newPath)
	}

	return path, true
}