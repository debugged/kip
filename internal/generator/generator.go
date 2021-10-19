package generator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

// DefaultGenerator for bla
const DefaultGenerator = "empty"

type generator struct {
	Name        string
	Info        string
	args        []string
	command     []string
	enableStdin bool
}

// Generators list
var Generators = []generator{
	generator{
		Name: DefaultGenerator,
		Info: "empty service with Dockerfile",
	},
	generator{
		Name: "nestjs",
		Info: "https://nestjs.com",
	},
	generator{
		Name: "angular",
		Info: "https://angular.io",
	},
	generator{
		Name: "react",
		Info: "https://reactjs.org",
	},
}

// Generate new project
func Generate(generatorName string, path string, name string, args []string) error {
	var buildErr error
	if len(generatorName) == 0 {
		generatorName = DefaultGenerator
	}

	switch generatorName {
	case "empty":
		buildErr = empty(path, name)
		break
	case "nestjs":
		buildErr = nestjs(path, name, args)
		break
	case "angular":
		buildErr = angular(path, name, args)
		break
	case "react":
		buildErr = react(path, name, args)
		break
	default:
		log.Fatalf("generator: \"%s\" does not exist\n", generatorName)
	}

	if buildErr != nil {
		return buildErr
	}

	return nil
}

func empty(path string, name string) error {
	servicePath := filepath.Join(path, name)

	err := os.MkdirAll(servicePath, os.ModePerm)

	emptyDockerfile := `FROM busybox
RUN echo "hello world!"`

	err = createDockerfile(servicePath, emptyDockerfile)

	if err != nil {
		return err
	}

	err = createDockerIgnore(servicePath, "")

	if err != nil {
		return err
	}

	return nil
}

// Generate nestjs
func nestjs(path string, name string, args []string) error {
	servicePath := filepath.Join(path, name)

	requireCommand("npx")
	cmdArgs := []string{"-p", "@nestjs/cli", "nest", "new", name, "--skip-git"}
	cmdArgs = append(cmdArgs, args...)
	err := runCommand(path, "npx", cmdArgs)

	if err != nil {
		return err
	}

	nestjsDockerIgnore := `node_modules
npm-debug.log
Dockerfile*
docker-compose*
.dockerignore
.git
.gitignore
README.md
LICENSE
.vscode`

	nestjsDockerfile := `FROM node:12 as base

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app
RUN chown -R node:node .

USER node

COPY --chown=node:node package.json .
COPY --chown=node:node package-lock.json* .


RUN npm ci
COPY --chown=node:node . .

FROM base as dev
CMD ["npm", "run", "start:dev"]

FROM base as build

WORKDIR /usr/src/app

RUN npm run build

FROM build
WORKDIR /usr/src/app

EXPOSE 3333

CMD [ "node", "dist/main.js" ]`

	err = createDockerfile(servicePath, nestjsDockerfile)

	if err != nil {
		return err
	}

	err = createDockerIgnore(servicePath, nestjsDockerIgnore)

	if err != nil {
		return err
	}

	return nil
}

func angular(path string, name string, args []string) error {
	servicePath := filepath.Join(path, name)

	requireCommand("npx")
	cmdArgs := []string{"-p", "@angular/cli", "ng", "new", name, "--skipGit=true"}
	err := runCommand(path, "npx", cmdArgs)

	if err != nil {
		return err
	}

	angularDockerIgnore := `node_modules
npm-debug.log
Dockerfile*
docker-compose*
.dockerignore
.git
.gitignore
README.md
LICENSE
.vscode`

	angularDockerfile := `FROM node:10 as modules

WORKDIR /usr/src/app
RUN chown -R node:node .

USER node

COPY --chown=node:node package.json .
COPY --chown=node:node package-lock.json .

RUN npm install

FROM node:10 as base
WORKDIR /usr/src/app

COPY --from=modules /usr/ /usr/
COPY . .

# Dev environment
FROM base as dev
CMD ["npm", "start"]

# We label our stage as ‘builder’
FROM base as builder

WORKDIR /usr/src/app

ARG configuration=production

## Build the angular app in production mode and store the artifacts in dist folder
RUN $(npm bin)/ng build --configuration $configuration

### STAGE 2: Setup ###

FROM nginx:1.15-alpine

## Copy our default nginx config
COPY nginx/default.conf /etc/nginx/conf.d/

## Remove default nginx website
RUN rm -rf /usr/share/nginx/html/*

## From ‘builder’ stage copy over the artifacts in dist folder to default nginx public folder
COPY --from=builder /usr/src/app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]`

	err = createDockerfile(servicePath, angularDockerfile)

	if err != nil {
		return err
	}

	err = createDockerIgnore(servicePath, angularDockerIgnore)

	if err != nil {
		return err
	}

	err = createNginConfig(servicePath)

	if err != nil {
		return err
	}

	return nil
}

func react(path string, name string, args []string) error {
	cmdArgs := []string{"create-react-app", name}
	cmdArgs = append(cmdArgs, args...)
	err := runCommand(path, "npx", cmdArgs)

	if err != nil {
		return err
	}

	return nil
}

func signalWatcher(ctx context.Context, cmd *exec.Cmd) {
	signalChan := make(chan os.Signal, 100)
	// Listen for all signals
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	signal := <-signalChan
	if err := cmd.Process.Signal(signal); err != nil {
		fmt.Println("Unable to forward signal: ", err)
	}
	for signal = range signalChan {
		if err := cmd.Process.Signal(signal); err != nil {
			fmt.Println("Unable to forward signal: ", err)
		}
	}
}

func runCommand(path string, command string, args []string) (err error) {
	requireCommand(command)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	go signalWatcher(ctx, cmd)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() != 0 {
					return errors.New("exited with non 0 code")
				}
				return nil
			}
		}
	}

	return nil
}

func requireCommand(command string) {
	_, err := exec.LookPath(command)
	if err != nil {
		log.Fatalf("command \"%s\" not found in $PATH\n", command)
	}
}

func createDockerfile(servicePath string, dockerfile string) error {
	f, err := os.Create(filepath.Join(servicePath, "Dockerfile"))
	if err != nil {
		return err
	}

	_, err = f.WriteString(dockerfile)
	if err != nil {
		return err
	}
	err = f.Close()

	if err != nil {
		return err
	}

	return nil
}

func createDockerIgnore(servicePath string, dockerIgnore string) error {
	f, err := os.Create(filepath.Join(servicePath, ".dockerignore"))
	if err != nil {
		return err
	}

	_, err = f.WriteString(dockerIgnore)
	if err != nil {
		return err
	}
	err = f.Close()

	if err != nil {
		return err
	}

	return nil
}

func createNginConfig(servicePath string) error {
	nginxPath := filepath.Join(servicePath, "nginx")

	err := os.MkdirAll(nginxPath, os.ModePerm)

	if err != nil {
		return err
	}

	nginxConfig := `server {

	listen 80;

	location / {
		root   /usr/share/nginx/html;
		index  index.html index.htm;
		try_files $uri $uri/ /index.html;
	}

	error_page   500 502 503 504  /50x.html;

	location = /50x.html {
		root   /usr/share/nginx/html;
	}

}`

	f, err := os.Create(filepath.Join(nginxPath, "default.conf"))
	if err != nil {
		return err
	}

	_, err = f.WriteString(nginxConfig)
	if err != nil {
		return err
	}
	err = f.Close()

	if err != nil {
		return err
	}

	return nil
}
