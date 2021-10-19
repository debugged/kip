package project

import (
	"debugged-dev/kip/v1/internal/version"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"robpike.io/filter"
)

// Project contains all the functions for maintaining a project
type Project interface {
	Name() string
	Version() string
	Template() string
	Environment() string
	Repository(enviroment string) string
	Paths() paths
	Charts() []Chart
	AddChart(chartName string, args []string) (string, error)
	Services() []ServiceProject
	GetService(name string) (*ServiceProject, error)
	GetScript(name string) (*Script, error)
	GetScripts(binding string, environment string) []Script
	AddScript(name string, command string, bindings []string) error
}

// MonoProject defined a project that contains multiple services
type MonoProject struct {
	path   string
	config *viper.Viper
}

func CreateMonoProject(path string, name string) error {
	project := MonoProject{path: filepath.Join(path, name)}
	return project.New(name)
}

type paths struct {
	Root              string
	Services          string
	Deployments       string
	Environments      string
	Scripts           string
	Libraries         string
	BuildPathTemplate string
}

type EnvConfig struct {
	Repository string `mapstructure:"repository"`
}

func (p MonoProject) Name() string {
	return filepath.Base(p.path)
}

func (p MonoProject) Template() string {
	return p.config.GetString("template")
}

func (p MonoProject) Environment() string {
	return p.config.GetString("environment")
}

func (p MonoProject) EnvConfig() map[string]*EnvConfig {
	envConfigs := make(map[string]*EnvConfig)
	p.config.UnmarshalKey("environments", &envConfigs)
	return envConfigs
}

func (p MonoProject) Repository(environment string) string {
	configs := p.EnvConfig()
	if val, ok := configs[environment]; ok {
		return val.Repository
	}
	return p.config.GetString("repository")
}

func (p MonoProject) Version() string {
	return p.config.GetString("version")
}

func (p MonoProject) Paths() paths {
	buildPathTemplate := "<serviceDir>"

	if p.config != nil && p.config.IsSet("buildPath") && len(p.config.GetString("buildPath")) > 0 {
		buildPathTemplate = p.config.GetString("buildPath")
	}

	return paths{
		Root:              p.path,
		Services:          filepath.Join(p.path, "services"),
		Deployments:       filepath.Join(p.path, "deployments"),
		Environments:      filepath.Join(p.path, "environments"),
		Scripts:           filepath.Join(p.path, "scripts"),
		Libraries:         filepath.Join(p.path, "libraries"),
		BuildPathTemplate: buildPathTemplate,
	}
}

func (p MonoProject) Services() []ServiceProject {
	return getServices(p.Paths().Services, &p)
}

func (p MonoProject) GetService(name string) (*ServiceProject, error) {
	services := p.Services()

	for _, service := range services {
		if service.Name() == name {
			return &service, nil
		}
	}

	return nil, fmt.Errorf("service %s not found", name)
}

func (p MonoProject) Charts() []Chart {
	return getCharts(p.Paths().Deployments, "", p)
}

func (p MonoProject) AddChart(chartName string, args []string) (string, error) {
	return createChart(chartName, p.Paths().Deployments, args)
}

func (p MonoProject) GetScript(name string) (*Script, error) {
	for _, script := range p.GetScripts("", "") {
		if script.Name == name {
			return &script, nil
		}
	}

	return nil, fmt.Errorf("script \"%s\" not found", name)
}

func (p MonoProject) GetScripts(binding string, environment string) []Script {
	var scripts []Script
	err := p.config.UnmarshalKey("scripts", &scripts)

	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	scripts = filter.Apply(scripts, func(s Script) Script {
		s.Path = p.Paths().Root
		return s
	}).([]Script)

	if binding != "" {
		scripts = filter.Choose(scripts, func(s Script) bool {
			for _, value := range s.Bindings {
				if value == binding {
					return true
				}
			}
			return false
		}).([]Script)
	}

	if environment != "" {
		scripts = filter.Choose(scripts, func(s Script) bool {

			if len(s.Environments) == 0 {
				return true
			}

			for _, value := range s.Environments {
				if value == environment {
					return true
				}
			}
			return false
		}).([]Script)
	}

	return scripts
}

func (p MonoProject) AddScript(scriptName string, command string, bindings []string) error {
	config := scriptConfig{Name: scriptName, Command: command, Bindings: bindings}

	scriptConfigs := []scriptConfig{}

	for _, script := range p.GetScripts("", "") {
		scriptConfigs = append(scriptConfigs, scriptConfig{Name: script.Name, Command: script.Command, Bindings: script.Bindings})
	}

	scriptConfigs = append(scriptConfigs, config)

	p.config.Set("scripts", scriptConfigs)

	err := p.config.WriteConfig()

	return err
}

func (p MonoProject) New(name string) error {
	paths := p.Paths()

	if _, err := os.Stat(paths.Root); !os.IsNotExist(err) {
		return fmt.Errorf("folder %s already exist", name)
	}

	os.MkdirAll(paths.Root, os.ModePerm)
	os.MkdirAll(paths.Services, os.ModePerm)
	os.MkdirAll(paths.Scripts, os.ModePerm)
	os.MkdirAll(paths.Environments, os.ModePerm)
	os.MkdirAll(paths.Deployments, os.ModePerm)

	config := viper.New()
	config.AddConfigPath(paths.Root)
	config.SetConfigName("kip_config")
	config.SetConfigType("yaml")

	config.Set("template", "project")
	config.Set("version", version.Get().Version)
	config.Set("environment", "dev")

	err := config.SafeWriteConfig()

	if err != nil {
		os.RemoveAll(paths.Root)
		return err
	}

	return nil
}

func (p MonoProject) Build(services []string, repository string, key string, args []string, environment string, debug bool) error {
	for _, service := range p.Services() {
		err := service.Build(repository, key, args, environment, debug)

		if err != nil {
			return err
		}
	}
	return nil
}

// GetProject creates the project class and makes it globally Available
func GetProject(projectPath string, config *viper.Viper) (Project, error) {
	switch config.GetString("template") {
	case "project":
		return MonoProject{path: projectPath, config: config}, nil
	case "service":
		return ServiceProject{path: projectPath, config: config}, nil
	default:
		return nil, fmt.Errorf("template %s not implemented", config.GetString("template"))
	}
}
