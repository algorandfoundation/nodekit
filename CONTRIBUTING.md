# 🫂 Contributing Guide

A guide on how to contribute to this project.

# Getting Started

Clone the project

```bash
git clone git@github.com:algorandfoundation/hack-tui.git
```

Change to the directory

```bash
cd hack-tui
```

Build the project

```bash
make build
```

Running a Participation Node

```bash
docker compose up
```

Create a configuration file for the participation node in the root directory of the project (.algorun.yaml)

```yaml
server: http://localhost:8080
token: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
```

Launch the TUI

```bash
./bin/algorun
```

# 📂 Folder Structure

```bash
├── api        # Generated API Client
├── cmd        # Command Controller
├── internal   # Data Models/Fetch Wrappers
└── ui         # BubbleTea Interfaces
```

There are three top level modules (**cmd**, **internal**, **ui**) which align with the GoLang/Charm ecosystem.
There is an additional code-generated module called **api** which should not be edited by hand.
See [generating rpc package](#generating-rpc-package) for more information

All submodules and endpoints **SHOULD** align with the command/ui namespaces.

Example Command:

```bash
hacktui status
```

Example Structure

```bash
├── cmd/status.go
├── internal/status.go
└── ui/status/table.go
```

All submodules **SHOULD** abstract when appropriate to a submodule.

Example Refactor

```bash
├── cmd/status/root.go
├── internal/status/model.go
└── ui/status/table.go
```

Additional top level modules **SHOULD NOT** be relied on unless there is a clear reason.
A common abstraction found in other projects is a `server` module which handles any local daemons.

### 🧑‍💻 cmd

Folder for commands that can be run from the cli.
Effectively this package is the "controller" in MVC

- **SHOULD** be used to manage cli commands
- **SHOULD** mirror the name of the command.
  `cli-tool command-name` should be represented as
  `./cmd/command-name.go` or `./cmd/command-name/root.go`
- **SHOULD** bind the `internal` and `ui` models
- **SHOULD NOT** contain any model or UI code (only bindings).

### 🪨 internal

Common library code which includes the models and business logic
of the application.
Its main responsibility is constructing the state used in the TUI.
This package is considered the "Model" in MVC

- **SHOULD** be used to hold models.
- **SHOULD** mirror the namespace the models relate to.
- **SHOULD NOT** contain any UI or CLI specific code (example, IsVisible or any tea|cobra interfaces).

### 💄 ui

Elements to be presented to the user.
This is built on the `bubbletea` abstraction.
This package is the ViewModel and View in MVC.

- **SHOULD** be used to build bubbletea interfaces.
- **SHOULD** be named by the component it represents.
  For example, `./ui/table.go` for a reusable component or
  `./ui/command-name/table.go` for context specific elements
- **SHOULD** contain ViewModel state like "IsVisible"
- **SHOULD NOT** contain any model or CLI specific code (ViewModels/tea.Models should be composed of internal Models for testability).

# Generating RPC package

The `api` package is generated via [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen).
Its configuration is found under `generate.yaml` and can be run with the following make command:

```bash
make generate
```

The full command for reference

```bash
oapi-codegen -config generate.yaml https://raw.githubusercontent.com/algorand/go-algorand/v3.26.0-stable/daemon/algod/api/algod.oas3.yml
```
