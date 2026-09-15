# Overview

Includes various [vhs](https://github.com/charmbracelet/vhs) tapes for the project.
Useful for creating consistent demos and guides when the TUI updates.

Tapes are recorded against a deterministic network staged by
[`tools/statewalker`](../tools/statewalker) (`statewalker stage demo`): a
private network with a long-validity online key, a near-expiry key, a zombie
key and an upgrade vote in progress. This replaces the old flow of typing
`nodekit` against whatever node the machine happened to run.

## Get Started

Requirements: Docker (for the pinned vhs renderer image) and `goal`/`algod`
on the `PATH` for the private network.

Record all tapes with:

```bash
make tapes
```

This builds `nodekit` and `statewalker`, stages the demo network
(`.tapes/setup.sh` writes its data dir to `.tapes/.datadir`), renders the
tapes with the pinned `ghcr.io/charmbracelet/vhs` container (`--network host`
so the tape's nodekit reaches the staged node; `DATADIR` is passed into the
tape shell), and tears the network down afterwards. Set `TAPE_SPEED=fast` to
use accelerated rounds instead of the MainNet-like default cadence.

The renderer is pinned via `VHS_IMAGE` in the Makefile so gifs stay
byte-stable across machines; a locally installed `vhs` also works if you
prefer (`cd .tapes && DATADIR=$(cat .datadir) vhs tui.tape`), but recent vhs
releases have silently rendered zero frames, hence the pin.

## Adding a tape

Copy the default `tui.tape` and name it appropriately:

```bash
cp ./tui.tape ./my-demo.tape
```

Edit the tape with your favorite editor (make sure to update the output
file), then attach to the staged network via `nodekit -d "$DATADIR"` like
`tui.tape` does. To iterate quickly without re-staging, run the setup once
and invoke vhs directly:

```bash
./.tapes/setup.sh
cd .tapes && DATADIR=$(cat .datadir) vhs ./my-demo.tape
```

### Theme

Example theme that uses some of the official Algorand Foundation brand guides

```
Set Theme { "name": "Whimsy", "black": "#2D2DFI", "red": "#ef6487", "green": "#5eca89", "yellow": "#fdd877", "blue": "#65aef7", "magenta": "#aa7ff0", "cyan": "#43c1be", "white": "#ffffff", "brightBlack": "#535178", "brightRed": "#ef6487", "brightGreen": "#5eca89", "brightYellow": "#fdd877", "brightBlue": "#65aef7", "brightMagenta": "#aa7ff0", "brightCyan": "#43c1be", "brightWhite": "#ffffff", "background": "#001324", "foreground": "#b3b0d6", "selection": "#3d3c58", "cursor": "#b3b0d6" }
```
