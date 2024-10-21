# Overview

The internal library holds the state machine and interfaces for the TUI. It largely is a wrapper around the
generated RPC client found in the api package. It supports gathering metrics from multiple sources, mainly
algod RPC and it's associated node.log file.
