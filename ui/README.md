# Overview

The ui package contains bubbletea interfaces. 

## Common practices

A `style.go` file holds lipgloss predefined styles for the package.

All components are instances of a `tea.Model` which is composed of models
from the `internal` package. 
Components can either be single file or independent packages.

Example for `status.go` single file component:

```go
package ui

import "github.com/algorandfoundation/algorun-tui/internal"

type StatusViewModel struct {
	Data internal.StateModel
	IsVisible bool
}

func (m StatusViewModel) Int(){}
//other tea.Model interfaces ...
```

Once the component is sufficiently complex or needs to be reused, it can be moved 
to its own package

Example refactor for `status.go` to a package:

#### ui/status/model.go
```go
package status
import "github.com/algorandfoundation/algorun-tui/internal"

type ViewModel struct {
	Data internal.StateModel
	IsVisible bool
}
```

#### ui/status/controller.go
```go
package status

// Init Lifecycle
func (m ViewModel) Init(){}

// Update lifecycle
func (m ViewModel) Update(){}
```

#### ui/status/view.go

```go
package status

func (m ViewModel) View() string {
	return "Amazing View"
}
```

#### ui/status/cmds.go

```go
package status

func EmitSomething(thing internal.Something) tea.Cmd {
	return func() tea.Msg {
		return thing
	}
}

```

#### ui/status/style.go

```go
package status

import "github.com/charmbracelet/lipgloss"

var someStyle = lipgloss.NewStyle()
```