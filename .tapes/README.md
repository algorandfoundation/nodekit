# Overview

Includes various [vhs](https://github.com/charmbracelet/vhs) tapes for the project. 
Useful for creating consistent demos and guides when the TUI updates


## Get Started

Start the test environment

```bash
docker compose up
```

Login to the container

```bash
docker exec -it --user nodekit test-tapes /bin/bash 
```

Edit the tapes with your favorite editor and output. 
Then you can run the vhs tape

```bash
vhs ./my-demo.tape
```

When you need to restart, you can bring the node down:

```bash
docker compose down
```

## CLI Tools

`./utils/*` contains scripts for `fnet` and automation. 
It includes an example runner `./utils/generate.sh` which can be used to run a suite of tapes on the instance.


## Tips

- All paths are relative to this directory (.tapes)
- Leverage the `./src/theme.tape` as a base:

    ```elixir
    Source ./src/theme.tape
    ```

- Artifacts are stored in ./artifacts
- The main node.run site can be used to test content
