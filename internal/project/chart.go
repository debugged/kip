package project

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Chart is used for describing a chart
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

func (c Chart) getHashes(environment string, args []string) (string, string, []string, string, error) {
	cmdArgs, files, err := getCommandArgsAndFiles(c, environment, args)
	if err != nil {
		return "", "", nil, "", err
	}

	osOpen := func(path string) (io.ReadCloser, error) {
		return os.Open(path)
	}

	dirHash, err := hash1(files, osOpen)
	if err != nil {
		return "", "", nil, "", err
	}

	commandHash, err := hashCommand(cmdArgs, dirHash)
	if err != nil {
		return "", "", nil, "", err
	}

	hashPath := filepath.Join(c.Project.Paths().Root, ".kip", "cache", "chart", c.name, environment+".hash")

	savedCommandHash, _ := getCommandHash(hashPath)

	return commandHash, savedCommandHash, cmdArgs, hashPath, nil
}

func (c Chart) IsChanged(environment string, args []string) (bool, error) {
	commandHash, savedCommandHash, _, _, err := c.getHashes(environment, args)

	if err != nil {
		return false, err
	}

	return commandHash != savedCommandHash, nil
}

func (c Chart) Deploy(environment string, args []string) error {
	commandHash, _, cmdArgs, hashPath, err := c.getHashes(environment, args)

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

	err = saveCommandHash(commandHash, hashPath)

	if err != nil {
		return err
	}

	return nil
}

func getCommandArgsAndFiles(c Chart, environment string, args []string) ([]string, []string, error) {
	cmdArgs := []string{"upgrade", c.Name(), ".", "--install"}
	files := []string{}

	if environment != "" {
		// check if values file is global
		if c.Project != nil {
			valuesfilePath := filepath.Join(c.Project.Paths().Environments, fmt.Sprintf("values-%s.yaml", environment))
			if _, err := os.Stat(valuesfilePath); err == nil {
				files = append(files, valuesfilePath)
				cmdArgs = append(cmdArgs, "-f", valuesfilePath)
			}

			if c.Project != nil && c.Project.Template() == "service" {
				valuesfilePath := filepath.Join(c.Project.(ServiceProject).Paths().Environments, fmt.Sprintf("values-%s.yaml", environment))
				if _, err := os.Stat(valuesfilePath); err == nil {
					files = append(files, valuesfilePath)
					cmdArgs = append(cmdArgs, "-f", valuesfilePath)
				}
			}
		}

		// check if values file is global
		valuesfilePath := filepath.Join(c.Path(), fmt.Sprintf("../values-%s.yaml", environment))
		if _, err := os.Stat(valuesfilePath); err == nil {
			files = append(files, valuesfilePath)
			cmdArgs = append(cmdArgs, "-f", valuesfilePath)
		}

		// check in chart folder
		valuesfilePath = filepath.Join(c.Path(), fmt.Sprintf("values-%s.yaml", environment))
		if _, err := os.Stat(valuesfilePath); err == nil {
			files = append(files, valuesfilePath)
			cmdArgs = append(cmdArgs, "-f", valuesfilePath)
		}
	}

	kipHelmArgs := strings.TrimSpace(os.Getenv("KIP_HELM_ARGS"))

	if kipHelmArgs != "" {
		helmArgs := strings.Split(kipHelmArgs, " ")
		cmdArgs = append(cmdArgs, helmArgs...)
	}

	cmdArgs = append(cmdArgs, args...)

	valueFileRegEx, e := regexp.Compile(c.Path() + "/values-.*.yaml")
	if e != nil {
		log.Fatal(e)
	}

	_ = filepath.Walk(c.Path(), func(path string, info os.FileInfo, err error) error {
		if err == nil && !valueFileRegEx.MatchString(path) && !info.IsDir() {
			files = append(files, path)
			return nil
		}
		return err
	})

	return cmdArgs, files, nil
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
			c := Chart{name: chartFolder, path: filepath.Join(path, chartFolder), prefix: prefix, Project: project}
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

func hashCommand(cmdArgs []string, dirHash string) (string, error) {
	h := sha256.New()

	h.Write([]byte(strings.Join(cmdArgs, " ")))
	h.Write([]byte(dirHash))

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func hash1(files []string, open func(string) (io.ReadCloser, error)) (string, error) {
	h := sha256.New()
	files = append([]string(nil), files...)
	sort.Strings(files)

	for _, file := range files {
		if strings.Contains(file, "\n") {
			return "", errors.New("dirhash: filenames with newlines are not supported")
		}
		r, err := open(file)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(h, r)
		r.Close()
		if err != nil {
			return "", err
		}
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

func getCommandHash(path string) (string, error) {
	b, err := ioutil.ReadFile(path) // just pass the file name
	if err != nil {
		return "", nil
	}

	return string(b), nil
}

func saveCommandHash(hash string, path string) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	_, err = f.WriteString(hash)

	if err := f.Close(); err != nil {
		return err
	}

	return err
}
