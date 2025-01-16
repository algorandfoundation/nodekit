## nodekit configure telemetry

Configure telemetry for the Algorand daemon

### Synopsis

                                                                                   
<img alt="Terminal Render" src="/assets/nodekit.png" width="65%">                            
                                                                                   
                                                                                   
Configure telemetry for the Algorand daemon                                        
                                                                                   
Overview:                                                                          
When a node is run using the algod command, before the script starts the server,   
it configures its telemetry based on the appropriate logging.config file.          
When a node’s telemetry is enabled, a telemetry state is added to the node’s logger
reflecting the fields contained within the appropriate config file                 
                                                                                   
The default telemetry provider is Nodely.                                          

```
nodekit configure telemetry [flags]
```

### Options

```
  -d, --datadir string    Data directory for the node
      --disable           Disables telemetry
      --enable            Enables telemetry
  -e, --endpoint string   Sets the "URI" property (default "https://tel.4160.nodely.io")
  -h, --help              help for telemetry
  -n, --name string       Enable Algorand remote logging with specified node name (default "anon")
```

### SEE ALSO

* [nodekit configure](/man/nodekit_configure.md)	 - Change settings on the system (WIP)

