# ⌨️ Hack-TUI

<div align="center">
    <img alt="Terminal Render" src="/assets/Banner.gif" width="65%">
</div>

<div align="center">
    <a target="_blank" href="https://github.com/algorandfoundation/hack-tui/actions/workflows/test.yaml">
        <img alt="CI Badge" src="https://github.com/algorandfoundation/hack-tui/actions/workflows/test.yaml/badge.svg"/>
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
Built with [bubbles](https://github.com/charmbracelet/bubbles) & [bubbletea](https://github.com/charmbracelet/bubbletea)

> [!CAUTION]
> This project is in alpha state and under heavy development. We do not recommend performing actions (e.g. key management) on participation nodes connected to public networks.

# 🚀 Get Started

Download the latest release by running

```bash
curl -fsSL https://nodekit.algorand.co/install.sh | bash
```

Launch the TUI by replacing the `<ENDPOINT>` and `<TOKEN>` 
with your server in the following example

> [!IMPORTANT]
> TUI requires the *admin* token in order to access participation key information. This can be found in the `algod.admin.token` file, e.g. `/var/lib/algorand/algod.admin.token`

```bash
algorun --algod-endpoint <ENDPOINT> --algod-token <TOKEN>
```

# ℹ️ Advanced Usage

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
## ⚙️ Configuration

Configuration is loaded in the following order:

1. [Command Line Flag Arguments](#flags)
2. [Configuration File](#configuration-file)
3. [Environment Variables](#environment-variables)
4. [ALGORAND_DATA Parsing](#algorand_data)

### Flags

The application supports the `algod-endpoint` and `algod-token` flags for configuration.

```bash
algorun --algod-endpoint http://localhost:8080 --algod-token aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
```

### Configuration File

The configuration file is named `.algorun.yaml` and is loaded in the following order:

1. Current Directory
2. Home Directory
3. /etc/algorun/

Example `.algorun.yaml` configuration file:

```yaml
algod-endpoint: "http://localhost:8080"
algod-token: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
```

### Environment Variables

Environment variables can be set in order to override a configuration or ALGORAND_DATA setting
but cannot be used to override the command line arguments.

The following are the additional ENV variables the TUI supports

| Name                   | Example                                                                                |
|------------------------|----------------------------------------------------------------------------------------|
| ALGORUN_ALGOD-ENDPOINT | ALGORUN_ALGOD-ENDPOINT="http://localhost:8080"                                         |
| ALGORUN_ALGOD-TOKEN    | ALGORUN_ALGOD-TOKEN="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" |

### ALGORAND_DATA

The TUI searches the environment for an `ALGORAND_DATA` variable. 
It then loads the `algod-token` and `algod-endpoint` values from
the algod data directory.

