package project

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Chart is used for describing a chart
type chart interface {
	Name() string
	Path() string
	Deploy(args []string) error
}

type Chart struct {
	name    string
	path    string
	prefix  string
	Project Project
}

func (c Chart) Name() string {
	if c.Prefix() != "" {
		return strings.Join([]string{c.prefix, c.name}, "-")
	}
	return c.name
}

func (c Chart) Path() string {
	return c.path
}

func (c Chart) Prefix() string {
	return c.prefix
}

func (c Chart) Deploy(environment string, args []string) error {
	fmt.Println(c.Path())

	cmdArgs := []string{"upgrade", c.Name(), ".", "--install"}

	if environment != "" {
		// check if values file is global
		if c.Project != nil {
			valuesfilePath := filepath.Join(c.Project.Paths().Environments, fmt.Sprintf("values-%s.yaml", environment))
			if _, err := os.Stat(valuesfilePath); err == nil {
				cmdArgs = append(cmdArgs, "-f", valuesfilePath)
			}

			if c.Project != nil && c.Project.Template() == "service" {
				valuesfilePath := filepath.Join(c.Project.(ServiceProject).Paths().Environments, fmt.Sprintf("values-%s.yaml", environment))
				if _, err := os.Stat(valuesfilePath); err == nil {
					cmdArgs = append(cmdArgs, "-f", valuesfilePath)
				}
			}
		}

		// check if values file is global
		valuesfilePath := filepath.Join(c.Path(), fmt.Sprintf("../values-%s.yaml", environment))
		if _, err := os.Stat(valuesfilePath); err == nil {
			cmdArgs = append(cmdArgs, "-f", valuesfilePath)
		}

		// check in chart folder
		valuesfilePath = filepath.Join(c.Path(), fmt.Sprintf("values-%s.yaml", environment))
		if _, err := os.Stat(valuesfilePath); err == nil {
			cmdArgs = append(cmdArgs, "-f", valuesfilePath)
		}
	}

	kipHelmArgs := strings.TrimSpace(os.Getenv("KIP_HELM_ARGS"))

	if kipHelmArgs != "" {
		helmArgs := strings.Split(kipHelmArgs, " ")
		cmdArgs = append(cmdArgs, helmArgs...)
	}

	cmdArgs = append(cmdArgs, args...)

	fmt.Printf("helm %s\n", strings.Join(cmdArgs, " "))

	cmd := exec.Command("helm", cmdArgs...)
	cmd.Dir = c.Path()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

func getCharts(path string, prefix string, project Project) []Chart {
	charts := []Chart{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return charts
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if f.IsDir() {
			chartFolder := f.Name()
			var c Chart
			c = Chart{name: chartFolder, path: filepath.Join(path, chartFolder), prefix: prefix, Project: project}
			charts = append(charts, c)
		}
	}

	return charts
}

func createChart(name string, path string, args []string) (string, error) {
	cmdArgs := []string{"create", name}
	cmd := exec.Command("helm", append(cmdArgs, args...)...)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return path, err
}
