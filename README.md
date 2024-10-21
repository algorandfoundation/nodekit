# ⌨️ Hack-TUI

<div align="center">
    <img alt="Terminal Render" src="/assets/Banner.gif" width="65%">
</div>

<div align="center">
    <a target="_blank" href="https://github.com/algorandfoundation/hack-tui">
        <img alt="CI Badge" src="https://img.shields.io/badge/CI-TODO-red"/>
    </a>
    <a target="_blank" href="https://github.com/algorandfoundation/hack-tui">
        <img alt="CD Badge" src="https://img.shields.io/badge/CD-TODO-red"/>
    </a>
    <a target="_blank" href="https://github.com/algorandfoundation/hack-tui/stargazers">
        <img alt="Repository Stars Badge" src="https://img.shields.io/github/stars/algorandfoundation/hack-tui?color=7B1E7A&logo=star&style=flat" />
    </a>
    <img alt="Repository Visitors Badge" src="https://api.visitorbadge.io/api/visitors?path=https%3A%2F%2Fgithub.com%2Falgorandfoundation%2Fhack-tui&countColor=%237B1E7A&style=flat" />
</div>

---

Terminal UI for managing Algorand nodes. 
Built with [bubbles](https://github.com/charmbracelet/bubbles)/[bubbletea](https://github.com/charmbracelet/bubbletea) 

# 🚀 Get Started

Run the build or ~~download the latest cli(WIP)~~.

## Building

Clone the repository

```bash
git clone git@github.com:algorandfoundation/hack-tui.git
```

Change to the project directory

```bash
cd hack-tui
```

Run the build command

```bash
make build
```

Start a participation node

```bash
docker compose up
```

Connect to the node 

```bash
./bin/algorun --server http://localhost:8080 --token aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
```

# ℹ️ Usage

## ⚙️ Configuration

Configuration is loaded in the following order:

1. Configuration file (.algorun.yaml)
   1. Current Directory
   2. Home Directory
   3. /etc/algorun/
2. ENV Configuration
   - ALGORUN_*
3. CLI Flag Arguments
4. ALGORAND_DATA parsing

This results in `ALGORAND_DATA` taking precedence in the loading order.

### .algorun.yaml

Example configuration file:

```yaml
server: "http://localhost:4001"
token: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
```

### Environment Variables

| Name           | Example                                                                          |
|----------------|----------------------------------------------------------------------------------|
| ALGORUN_SERVER | ALGORUN_SERVER="http://localhost:4001"                                           |
| ALGORUN_TOKEN  | ALGORUN_TOKEN="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" |


### Flags

The application supports the `server` and `token` flags for configuration.

```bash
algorun --server http://localhost:4001 --token aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
```

## 🧑‍💻 Commands

The default command will launch the full TUI application

```bash
algorun
```

### Status

Render only the status overview in the terminal

```bash
algorun status
```

### Help

Display the usage information for the command

```bash
algorun help
```