package cmdshell

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type shell interface {
	Run() string
	WorkingDir(path string) error
}

type Shell struct {
	cmd *exec.Cmd
}

func CreateShell(shell string) (CmdShell, error) {
	cmd := exec.Command("zsh")
	cmd.Dir = filepath.Join(s.Path, "..")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()

	if err != nil {
		return err
	}

	err = cmd.Start()

	if err != nil {
		return err
	}

	_, err = io.WriteString(stdin, "echo test\n")

	if err != nil {
			log.Fatal(err)
	}

	_, err = io.WriteString(stdin, "echo bla\n")
	
	if err != nil {
			log.Fatal(err)
	}

	return Shell{cmd: cmd}, nil
} 