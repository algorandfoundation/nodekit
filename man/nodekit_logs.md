## nodekit logs

Display the node's logs

### Synopsis

                                                                             
<img alt="Terminal Render" src="/assets/nodekit.png" width="65%">                      
                                                                             
                                                                             
Display the node's logs                                                      
                                                                             
Overview:                                                                    
Reads the log file algod writes, wherever config.json puts it.               
Only warnings and errors are shown by default; use --all to see every level. 
Lines that are not valid log entries, such as crash output, are always shown.
                                                                             
Notes:                                                                       
Every matching entry is shown; --lines N shows only the newest N of them.    
--lines counts entries that match the filters, not raw lines of the file.    
--follow shows the newest 10 before it starts streaming, the way tail -f     
does; --lines N sets that backlog, and --lines 0 shows the whole history.    
The rotated archives are read as well, so the history reaches back past the  
last rotation. Pass --file to read one file on its own instead.              
--filter matches plain text in the message and the fields shown beside it,   
such as Round=49291042, with no pattern syntax.                              
A node only writes entries at or above its configured level, so asking for a 
lower level than the node records will find nothing.                         
                                                                             

```
nodekit logs [flags]
```

### Options

```
  -a, --all              Show entries at every level
  -d, --datadir string   Data directory for the node
  -F, --file string      Read this log file instead of the node's own
      --filter string    Only entries whose message or shown fields contain this text
  -f, --follow           Stream new entries as they are written
  -h, --help             help for logs
      --json             Emit the raw JSON log entries
      --level string     Minimum level to show: trace, debug, info, warn, error, fatal, panic
  -n, --lines int        Number of newest matching entries to show, 0 for all (10 with --follow)
      --since string     Only entries newer than a duration (15m, 2h) or a timestamp
```

### SEE ALSO

* [nodekit](/README.md)	 - Manage Algorand nodes from the command line

