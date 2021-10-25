package project

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Chart is used for describing a chart
type Chart struct {
	name    string
	path    string
	Project Project
}

func (c Chart) Name() string {
	return c.name
}

func (c Chart) Path() string {
	return c.path
}

func (c Chart) getHashes(environment string, args []string) (string, string, error) {
	cmdArgs, err := getCommandArgsAndFiles(c, environment, args, true)
	if err != nil {
		return "", "", err
	}

	cmd := exec.Command("helm", cmdArgs...)
	cmd.Dir = c.Path()

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println(string(output))
		log.Fatal(err)
		return "", "", err
	}

	h := sha256.New()
	h.Write(output)
	commandHash := base64.StdEncoding.EncodeToString(h.Sum(nil))

	savedHash, _ := getSavedHash("kip." + c.Name())

	return commandHash, savedHash, nil
}

func (c Chart) IsChanged(environment string, args []string) (bool, error) {
	commandHash, savedCommandHash, err := c.getHashes(environment, args)

	if err != nil {
		return false, err
	}

	return commandHash != savedCommandHash, nil
}

func (c Chart) Deploy(environment string, args []string) error {
	commandHash, _, err := c.getHashes(environment, args)

	if err != nil {
		return err
	}

	cmdArgs, err := getCommandArgsAndFiles(c, environment, args, false)
	if err != nil {
		return err
	}

	fmt.Printf("helm %s\n", strings.Join(cmdArgs, " "))

	cmd := exec.Command("helm", cmdArgs...)
	cmd.Dir = c.Path()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		return err
	}

	err = saveCommandHash(commandHash, "kip."+c.Name())

	if err != nil {
		return err
	}

	return nil
}

func getCommandArgsAndFiles(c Chart, environment string, args []string, template bool) ([]string, error) {
	cmdArgs := []string{}

	if template {
		cmdArgs = append(cmdArgs, []string{"template", c.Name(), "."}...)
	} else {
		cmdArgs = append(cmdArgs, []string{"upgrade", c.Name(), ".", "--install"}...)
	}

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

	return cmdArgs, nil
}

func getCharts(path string, project Project) []Chart {
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
			c := Chart{name: chartFolder, path: filepath.Join(path, chartFolder), Project: project}
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

func getSavedHash(secretName string) (string, error) {
	cmdArgs := []string{"get", "secret", secretName, "-o", "jsonpath={.data.hash}"}
	cmd := exec.Command("kubectl", cmdArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(string(output))
	if err != nil {
		log.Fatal(err)
	}

	return string(decoded), nil
}

func saveCommandHash(hash string, secretName string) error {
	cmdArgs := []string{"delete", "secret", secretName}
	cmd := exec.Command("kubectl", cmdArgs...)
	cmd.Run()

	cmdArgs = []string{"create", "secret", "generic", secretName, "--from-literal=hash=" + hash}
	cmd = exec.Command("kubectl", cmdArgs...)
	err := cmd.Run()

	return err
}
