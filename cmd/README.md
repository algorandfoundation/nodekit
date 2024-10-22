# Overview

The `cmd` package is used as the entrypoint for the end users. 
It is built using `cobra` and `viper`. 

Commands are largely responsible for binding the internal models to the TUI models.
This includes any state events between the two packages (`ui` and `internal`)

## RootCMD (root.go)

- The main Execute method which contains all subcommands. 
- Bootstraps the api and configurations.
- Mounts the Viewport as the default command

## Status (status.go)

- Renders the Status component