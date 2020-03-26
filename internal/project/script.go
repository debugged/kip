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

type script interface {
	Run(args []string) error
}

type scriptConfig struct {
	Name string
	Command string
	Bindings []string
	Args []string
}

type Script struct {
	Name string
	Command string
	Path string
	Bindings []string
	Args []string
}

func (s Script) Run(args []string) error {
	fmt.Println(s.Path)

	cmdArgs := s.Args
	cmdArgs = append(cmdArgs, args...)

	fmt.Printf("%s %s\n", s.Command, strings.Join(cmdArgs, " "))

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
		return err;
	}


	lines := strings.Split(stdBuffer.String(), "\n")
	
	r := regexp.MustCompile(`^(?P<key>[A-z0-9]*)(=)(?P<value>.*)$`)

	for _, line := range lines {
		if r.MatchString(line) {
			result := strings.Split(line, "=")
			os.Setenv(result[0], result[1])
		}
	}

	return nil
}

func getScripts(path string) []Script {
	scripts := []Script{}
	return scripts
}