# Overview

Includes various [vhs](https://github.com/charmbracelet/vhs) tapes for the project. 
Useful for creating consistent demos and guides when the TUI updates


## Get Started

Install vhs

```bash
go install github.com/charmbracelet/vhs@latest
```

Copy the default `tui.tape` and name it appropriately

```bash
cp ./tui.tape ./my-demo.tape
```

Edit the tape with your favorite editor. 
Then you can run the vhs tape

(Make sure to update the output file)

```bash
vhs ./my-demo.tape
```

### Theme

Example theme that uses some of the official Algorand Foundation brand guides

```
Set Theme { "name": "Whimsy", "black": "#2D2DFI", "red": "#ef6487", "green": "#5eca89", "yellow": "#fdd877", "blue": "#65aef7", "magenta": "#aa7ff0", "cyan": "#43c1be", "white": "#ffffff", "brightBlack": "#535178", "brightRed": "#ef6487", "brightGreen": "#5eca89", "brightYellow": "#fdd877", "brightBlue": "#65aef7", "brightMagenta": "#aa7ff0", "brightCyan": "#43c1be", "brightWhite": "#ffffff", "background": "#001324", "foreground": "#b3b0d6", "selection": "#3d3c58", "cursor": "#b3b0d6" }
```

