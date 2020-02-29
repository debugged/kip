package project

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Service is used for describing a service
type Service interface {
	Name() string
	Path() string
	HasDockerfile() bool
	Build(args []string) error
}

type service struct {
	name string
	path string
}


func (s service) Name() string { 
	return s.name
}

func (s service) Path() string { 
	return s.path
}

func (s service) HasDockerfile() bool { 
	_, err := os.Stat(filepath.Join(s.path, "Dockerfile"))
	return !os.IsNotExist(err)
}

func (s service) Build(args []string) error {
	fmt.Println(s.Path())
	cmdArgs := []string{"build", ".", "-t", s.Name()}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Dir = s.Path()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		fmt.Fprintln(os.Stdout, err)
	}
	
	return err
}

func getServices(path string) []Service {
	services := []Service{}

	files, err := ioutil.ReadDir(path)
			if err != nil {
					log.Fatal(err)
			}
			

			for _, f := range files {
				if f.IsDir() {
					serviceName := f.Name()

					var s service
					s = service{name: serviceName, path: filepath.Join(path, serviceName)}
					services = append(services, s)
				}
			}

	return services
}