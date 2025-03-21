## nodekit catchup start

Get the latest catchpoint and start catching up.

### Synopsis

                                                                                                 
<img alt="Terminal Render" src="/assets/nodekit.png" width="65%">                                          
                                                                                                 
                                                                                                 
Catchup the node to the latest catchpoint.                                                       
                                                                                                 
Overview:                                                                                        
Starting a catchup will sync the node to the latest catchpoint.                                  
Actual sync times may vary depending on the number of accounts, number of blocks and the network.
                                                                                                 
Note: Not all networks support Fast-Catchup.                                                     

```
nodekit catchup start [flags]
```

### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for start
```

### SEE ALSO

* [nodekit catchup](/man/nodekit_catchup.md)	 - Manage Fast-Catchup for your node

