# ℹ️ Overview

Following in the footsteps of [node-ui](https://github.com/algorand/node-ui), this project
should be built as a first class TUI interface for Algorand daemons.

## ✅ Decisions

- **SHOULD** be built with [Charm](https://charm.sh/)/[Bubbletea](https://github.com/charmbracelet/bubbletea) and [Bubbles](https://github.com/charmbracelet/bubbles).
- **SHOULD** conform to GoLang, Charm, and Bubbles practices of other TUI applications
- **SHOULD** maintain high level of unit testing

## 🔨 Deliverables

- facilitate key registration
- thin algod REST client
- display status and health