package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type script interface {
	Run(args []string) error
}

type scriptConfig struct {
	Name string
	Command string
	Bindings []string
}

type Script struct {
	Name string
	Command string
	Path string
	Bindings []string
}

func (s Script) Run(args []string) error {
	fmt.Println(s.Path)

	cmdArgs := []string{s.Name}
	cmdArgs = append(cmdArgs, args...)

	fmt.Printf("%s %s\n", s.Command, strings.Join(cmdArgs, " "))

	cmd := exec.Command(s.Command, cmdArgs...)
	cmd.Dir = filepath.Join(s.Path, "..")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	output, err := cmd.Output()

	if err != nil {
		fmt.Println(err)
		return err;
	}

	outputString := string(output)

	lines := strings.Split(outputString, "\n")
	
	r := regexp.MustCompile(`^(?P<key>.*)(=)(?P<value>.*)$`)

	if 1 > 2 {
		for _, line := range lines {
			r.MatchString(line)
		}
	}

	return nil
}

func getScripts(path string) []Script {
	scripts := []Script{}
	return scripts
}

// func createChart(name string, path string, args []string) (string, error) {
// 	cmdArgs := []string{"create", name}
// 	cmd := exec.Command("helm", append(cmdArgs, args...)...)
// 	cmd.Dir = path
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	err := cmd.Run()
// 	return path, err
