package project

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type scriptConfig struct {
	Name         string
	Command      string
	Bindings     []string
	Args         []string
	Environments []string
}

type Script struct {
	Name         string
	Command      string
	Path         string
	Bindings     []string
	Args         []string
	Environments []string
}

func (s Script) Run(args []string) error {
	cmdArgs := s.Args
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(s.Command, cmdArgs...)
	cmd.Dir = s.Path
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw

	err := cmd.Run()

	if err != nil {
		fmt.Println(err)
		return err
	}

	lines := strings.Split(stdBuffer.String(), "\n")

	r := regexp.MustCompile(`^(?P<key>[A-z0-9]*)(=)(?P<value>.*)$`)

	for _, line := range lines {
		if r.MatchString(line) {
			matches := r.FindStringSubmatch(line)

			if matches[1] == "KIP_HELM_ARGS" {
				os.Setenv("KIP_HELM_ARGS", fmt.Sprintf("%s %s", os.Getenv("KIP_HELM_ARGS"), matches[3]))
			} else {
				os.Setenv(matches[1], matches[3])
			}

		}
	}

	return nil
}

func getScripts(path string) []Script {
	scripts := []Script{}
	return scripts
}
