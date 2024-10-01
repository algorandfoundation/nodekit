# 🫂 Contributing Guide [WIP]

## Getting Started

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
go build .
```


## 📂 Folder Structure [WIP]

There are three top level modules (**cmd**, **internal**, **ui**) which align with the GoLang/Charm ecosystem.

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
├── cmd/status/main.go
├── internal/status/main.go
└── ui/status/table.go
```

Additional top level modules **SHOULD NOT** be relied on unless there is a clear reason.
A common abstraction found in other projects is a `server` module which handles any local daemons.

### 🧑‍💻 cmd

Folder for commands that can be run from the cli.

- **SHOULD** be used to manage cli commands
- **SHOULD** mirror the name of the command.
  `cli-tool command-name` should be represented as
  `./cmd/command-name.go` or `./cmd/command-name/main.go`
- **SHOULD NOT** contain any model or UI code.

### 🪨 internal

Common library code which includes the models and business logic
of the application

- **SHOULD** be used to hold models.
- **SHOULD** mirror the namespace the models relate to.
- **SHOULD NOT** contain any UI or CLI specific code.

### 💄 ui

Elements to be presented to the user.

- **SHOULD** be used to build bubbles interfaces.
- **SHOULD** be named by the component it represents.
  For example, `./ui/table.go` for a reusable component or
  `./ui/command-name/table.go` for context specific elements
- **SHOULD NOT** contain any model or CLI specific code.
