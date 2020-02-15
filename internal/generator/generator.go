package generator

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// DefaultGenerator for bla
const DefaultGenerator = "empty"

type generator struct {
	Name string
	Info string
	args []string
	command []string
	enableStdin bool
}
// Generators list
var Generators = []generator {
	generator {
		Name: DefaultGenerator,
		Info: "empty service with Dockerfile",
	},
	generator {
		Name: "nestjs",
		Info: "https://nestjs.com",
	},
}

// Generate new project
func Generate(generatorName string, path string, name string)  {
	if len(generatorName) == 0 { 
		generatorName = DefaultGenerator
	}

	switch generatorName {
	case "empty":
		empty(path, name)
		break
	case "nestjs":
		nestjs(path, name)
		break
	default:
		fmt.Printf(`generator: "%s" does not exist`,generatorName)
		os.Exit(1)
	}
	fmt.Printf(`project: %s generated in: %s`, name, path)
}

func empty(path string, name string) {

}

// Generate nestjs
func nestjs(path string, name string) {
	cmd := exec.Command("npx", "-p", "@nestjs/cli", "nest", "new", name)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	
	if err != nil {
		log.Fatal(err)
	}

	// todo cleanup if command exit code =! 0
}