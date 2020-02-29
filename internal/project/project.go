package project

import (
	"path/filepath"

	"github.com/spf13/viper"
)

// Project contains all the functions for maintaining a project
type Project interface {
	Name() string
	Version() string
	Paths() paths
	Services() []Service
}

type project struct {
	path string
	config *viper.Viper
}

type paths struct {
	Root string
	Services string
	Deployments string
}

func (p project) Name() string { 
	return p.config.GetString("name")
}

func (p project) Version() string { 
	return p.config.GetString("version")
}


func (p project) Paths() paths { 
	return paths { 
		Root: p.path, 
		Services: filepath.Join(p.path, "services"), 
		Deployments: filepath.Join(p.path, "deployment"),
	}
}

func (p project) Services() []Service  {
	return getServices(p.Paths().Services)
}

// GetProject creates the project class and makes it globally Available
func GetProject(projectPath string, config *viper.Viper) Project {
	return project{path: projectPath, config: config}
}