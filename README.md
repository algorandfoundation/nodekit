## nodekit

Manage Algorand nodes from the command line

### Synopsis

                                                                                                    
<img alt="Terminal Render" src="/assets/nodekit.png" width="65%">                                             
                                                                                                    
                                                                                                    
Manage Algorand nodes from the command line                                                         
                                                                                                    
Overview:                                                                                           
Welcome to NodeKit, a TUI for managing Algorand nodes.                                              
A one stop shop for managing Algorand nodes, including node creation, configuration, and management.
                                                                                                    
Note: This is still a work in progress. Expect bugs and rough edges.                                

```
nodekit [flags]
```

### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for nodekit
  -n, --no-incentives    Disable setting incentive eligibility fees
```

### SEE ALSO

* [nodekit bootstrap](/man/nodekit_bootstrap.md)	 - Initialize a fresh node
* [nodekit catchup](/man/nodekit_catchup.md)	 - Manage Fast-Catchup for your node
* [nodekit configure](/man/nodekit_configure.md)	 - Change settings on the system (WIP)
* [nodekit debug](/man/nodekit_debug.md)	 - Display debugging information
* [nodekit install](/man/nodekit_install.md)	 - Install the node daemon
* [nodekit start](/man/nodekit_start.md)	 - Start the node daemon
* [nodekit stop](/man/nodekit_stop.md)	 - Stop the node daemon
* [nodekit telemetry](/man/nodekit_telemetry.md)	 - NoOp command
* [nodekit uninstall](/man/nodekit_uninstall.md)	 - Uninstall the node daemon
* [nodekit upgrade](/man/nodekit_upgrade.md)	 - Upgrade the node daemon


### Installing

Connect to your server and run the installation script which will bootstrap your node.

```bash
curl -fsSL https://nodekit.run/install.sh | bash
```
