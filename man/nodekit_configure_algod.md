## nodekit configure algod

Configure algod for the Algorand daemon

### Synopsis


<img alt="Terminal Render" src="/assets/nodekit.png" width="65%">


Configure algod for the Algorand daemon

Overview:
When nodekit is run using the configure algod command, it displays the current
state of the algod config.json file that NodeKit allows you to modify. When
flags are provided to the configure algod command, then the config.json file is
updated to reflect the chosen options. Once a configuration has taken place,
NodeKit will attempt to restart the algod process. Algod must be restarted
before the changes take place.

```
nodekit configure algod [flags]
```

### Options

```
  -d, --datadir string   Data directory for the node
  -h, --help             help for algod
      --hybrid           Enable or Disable P2P Hybrid Mode (default true)
```

### SEE ALSO

* [nodekit configure](/man/nodekit_configure.md)	 - Change settings on the system (WIP)

