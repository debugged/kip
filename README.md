<h1 align="center">Welcome to Kip 🐔</h1>
<p>
  <img alt="Version" src="https://img.shields.io/badge/version-V0.0.1--alpha-blue.svg?cacheSeconds=2592000" />
  <a href="#" target="_blank">
    <img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg" />
  </a>
</p>

> Kip is an easy to use CLI for developing and deploying your kubernetes project.

# **Notice**

**Kip is under active development. We're open to suggestions and feedback, please feel free to open an issue**

# Description

TODO

# Prerequisites

- kubectl ([more info](https://github.com/kubernetes/kubectl))
- helm ([more info](https://github.com/helm/helm))
- docker ([more info](https://docs.docker.com/install/))

# Install

## from source

```bash
make install
```

## from release

Downwload binary from: [releases](/releases)

### Linux

```bash
mv kip ~/.local/bin/kip
chmod +x ~/.local/bin/kip
```

### Mac

```bash
mv kip /usr/local/bin/kip
chmod +x ~/.local/bin/kip
```

### Windows

1. Download binary
2. Add the binary to your PATH.

# Commands

| Command                 | Description                                           |
| ----------------------- | ----------------------------------------------------- |
| `kip build`             | Builds one or more services                           |
| `kip chart add`         | Adds a new helm chart to your project or service      |
| `kip chart list`        | Lists all charts                                      |
| `kip check`             | Checks if all dependencies are in available in \$PATH |
| `kip deploy`            | Deploys project or service                            |
| `kip generators`        | Lists all available generators for creating services  |
| `kip help`              | List all available commands                           |
| `kip new [NAME]`        | Creates a new kip project                             |
| `kip run [SCRIPT_NAME]` | Runs a script                                         |
| `kip script add`        | Add a new script to your project or service           |
| `kip script list`       | List all scripts                                      |
| `kip service add`       | Create a new service                                  |
| `kip service list`      | List all services                                     |
| `kip version`           | Print the client version information                  |

### kip new

This is the initialization command that creates a new project or service.
Example:

```bash
kip new foobar
```

The following (optional) flags are available:

```
* -t, --template string        project | service. A project can contain multiple services(default "project")
* -g, --generator string       Generator used for creating service projects (ex:nestjs,angular)
* -h, --help                   Extra information about the kip new command
```

# Usage

### 1. Create project

TODO
