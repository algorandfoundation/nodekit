# ℹ️ Overview

Node Management has many aspects which this decision makes concrete.

## ✅ Decisions

- **SHOULD** include install/upgrade commands  
- **SHOULD** include start/stop commands
- **SHOULD** include catchup commands
- **SHOULD** include bootstrap command

## 🔨 Deliverables

- Use package managers for installation and upgrades (brew, dnf, apt-get)
- Use native supervisors for algod orchestration (launchd, systemd)
- Bootstrap concept which ties several components together (install, start, fast-catchup, launch TUI)
- Limited amount of configurations for the initial release