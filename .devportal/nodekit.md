---
title: "nodekit"
---
## Synopsis                                             
                                                                                                    
                                                                                                    
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

# Commands 
## bootstrap                                                  
                                                                                                         
                                                                                                         
Initialize a fresh node                                                                                  
                                                                                                         
Overview:                                                                                                
Get up and running with a fresh Algorand node.                                                           
Uses the local package manager to install Algorand, and then starts the node and preforms a Fast-Catchup.
                                                                                                         
Note: This command only supports the default data directory, /var/lib/algorand                           

```
nodekit bootstrap [flags]
```

#### Options

```
  -h, --help   help for bootstrap
```

## catchup                                          
                                                                                                 
                                                                                                 
Fast-Catchup is a feature that allows your node to catch up to the network faster than normal.   
                                                                                                 
Overview:                                                                                        
The entire process should sync a node in minutes rather than hours or days.                      
Actual sync times may vary depending on the number of accounts, number of blocks and the network.
                                                                                                 
Note: Not all networks support Fast-Catchup.                                                     

```
nodekit catchup [flags]
```

#### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for catchup
```

## catchup debug      
                                                             
                                                             
Display debug information for Fast-Catchup.                  
                                                             
Overview:                                                    
This information is useful for debugging fast-catchup issues.
                                                             
Note: Not all networks support Fast-Catchup.                 

```
nodekit catchup debug [flags]
```

#### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for debug
```

## catchup start                                          
                                                                                                 
                                                                                                 
Catchup the node to the latest catchpoint.                                                       
                                                                                                 
Overview:                                                                                        
Starting a catchup will sync the node to the latest catchpoint.                                  
Actual sync times may vary depending on the number of accounts, number of blocks and the network.
                                                                                                 
Note: Not all networks support Fast-Catchup.                                                     

```
nodekit catchup start [flags]
```

#### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for start
```

## catchup stop                            
                                                                                   
                                                                                   
Stop a fast catchup                                                                
                                                                                   
Overview:                                                                          
Stop an active Fast-Catchup. This will abort the catchup process if one has started
                                                                                   
Note: Not all networks support Fast-Catchup.                                       

```
nodekit catchup stop [flags]
```

#### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for stop
```

## configure                             
                                                                                    
                                                                                    
Change settings on the system (WIP)                                                 
                                                                                    
Overview:                                                                           
Tool for inspecting and updating the Algorand daemon's config.json and service files
                                                                                    
Note: This is still a work in progress. Expect bugs and rough edges.                

#### Options

```
  -h, --help   help for configure
```

## configure service                 
                                                                        
                                                                        
Install service files for the Algorand daemon.                          
                                                                        
Overview:                                                               
Ensuring that the Algorand daemon is installed and running as a service.
                                                                        
Note: This is still a work in progress. Expect bugs and rough edges.    

```
nodekit configure service [flags]
```

#### Options

```
  -h, --help   help for service
```

## configure telemetry                            
                                                                                   
                                                                                   
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

#### Options

```
  -d, --datadir string    Data directory for the node
      --disable           Disables telemetry
      --enable            Enables telemetry
  -e, --endpoint string   Sets the "URI" property (default "https://tel.4160.nodely.io")
  -h, --help              help for telemetry
  -n, --name string       Enable Algorand remote logging with specified node name (default "anon")
```

## debug                               
                                                                                      
                                                                                      
Display debugging information                                                         
                                                                                      
Overview:                                                                             
Prints the known state of the nodekit                                                 
Checks various paths and configurations to present useful information for bug reports.
                                                                                      

```
nodekit debug [flags]
```

#### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for debug
```

## install                                    
                                                                                           
                                                                                           
Install the node daemon                                                                    
                                                                                           
Overview:                                                                                  
Configures the local package manager and installs the algorand daemon on your local machine
                                                                                           

```
nodekit install [flags]
```

#### Options

```
  -f, --force   forcefully install the node
  -h, --help    help for install
```

## start                                                                        
                                                                                                                               
                                                                                                                               
Start the node daemon                                                                                                          
                                                                                                                               
Overview:                                                                                                                      
Start the Algorand daemon on your local machine if it is not already running. Optionally, the daemon can be forcefully started.
                                                                                                                               
This requires the daemon to be installed on your system.                                                                       

```
nodekit start [flags]
```

#### Options

```
  -f, --force   forcefully start the node
  -h, --help    help for start
```

## stop                                           
                                                                                                  
                                                                                                  
Stop the node daemon                                                                              
                                                                                                  
Overview:                                                                                         
Stops the Algorand daemon on your local machine. Optionally, the daemon can be forcefully stopped.
                                                                                                  
This requires the daemon to be installed on your system.                                          

```
nodekit stop [flags]
```

#### Options

```
  -f, --force   forcefully stop the node
  -h, --help    help for stop
```

## telemetry 
                                                        
                                                        
NoOp command                                            
                                                        
Overview:                                               
Lorem ipsum dolor sit amet, consectetur adipiscing elit.
                                                        

#### Options

```
  -h, --help   help for telemetry
```

## telemetry disable             
                                                                    
                                                                    
Disable Telemetry                                                   
                                                                    
Overview:                                                           
Configure telemetry for the Algorand daemon.                        
                                                                    
Note: This is still a work in progress. Expect bugs and rough edges.

#### Options

```
  -h, --help   help for disable
```

## telemetry                                                                                                                      
                                                                                                                                                                             
                                                                                                                                                                             
Disable Nodely                                                                                                                                                               
                                                                                                                                                                             
Overview:                                                                                                                                                                    
Nodely Telemetry is a free telemetry service offered by a third party (Nodely)                                                                                               
Enabling telemetry will configure your node to send health metrics to Nodely                                                                                                 
Privacy note: Information about your node (including participating accounts and approximate geographic location) will be associated with an anonymous user identifier (GUID.)
Tip: Keep this GUID identifier private if you do not want this information to be linked to your identity.                                                                    
                                                                                                                                                                             
Note: This is still a work in progress. Expect bugs and rough edges.                                                                                                         

```
nodekit telemetry disable nodely [flags]
```

#### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for nodely
```

## telemetry enable             
                                                                    
                                                                    
Enable Telemetry                                                    
                                                                    
Overview:                                                           
Configure telemetry for the Algorand daemon.                        
                                                                    
Note: This is still a work in progress. Expect bugs and rough edges.

#### Options

```
  -h, --help   help for enable
```

## telemetry             
                                                                    
                                                                    
Disable Nodely                                                      
                                                                    
Overview:                                                           
Configure telemetry for the Algorand daemon.                        
                                                                    
Note: This is still a work in progress. Expect bugs and rough edges.

```
nodekit telemetry enable nodely [flags]
```

#### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for nodely
```

## telemetry status             
                                                                    
                                                                    
Status of telemetry                                                 
                                                                    
Overview:                                                           
Configure telemetry for the Algorand daemon.                        
                                                                    
Note: This is still a work in progress. Expect bugs and rough edges.

```
nodekit telemetry status [flags]
```

#### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for status
```

## uninstall                                  
                                                                                         
                                                                                         
Uninstall the node daemon                                                                
                                                                                         
Overview:                                                                                
Uninstall Algorand node (Algod) and other binaries on your system installed by this tool.
                                                                                         
This requires the daemon to be installed on your system.                                 

```
nodekit uninstall [flags]
```

#### Options

```
  -f, --force   forcefully uninstall the node
  -h, --help    help for uninstall
```

## upgrade            
                                                                   
                                                                   
Upgrade the node daemon                                            
                                                                   
Overview:                                                          
Upgrade Algorand packages if it was installed with package manager.
                                                                   
This requires the daemon to be installed on your system.           

```
nodekit upgrade [flags]
```

#### Options

```
  -h, --help   help for upgrade
```

###### Auto generated by spf13/cobra on 13-Feb-2025
