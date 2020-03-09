package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Start()

	if err != nil {
		fmt.Println(err)
		return err;
	}

	// cmd := exec.Command("bash ./script.sh")
	// cmd.Dir = filepath.Join(s.Path, "..")
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// cmd.Env = []string{"MY_VAR=some_value"}

	// stdin, err := cmd.StdinPipe()

	// if err != nil {
	// 	return err
	// }

	// err := cmd.Start()

	// if err != nil {
	// 	return err
	// }

	// _, err = io.WriteString(stdin, "echo test\n")

	// if err != nil {
	// 		log.Fatal(err)
	// }

	// _, err = io.WriteString(stdin, "echo bla\n")
	
	// if err != nil {
	// 		log.Fatal(err)
	// }

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
