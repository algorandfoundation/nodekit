package main

import (
	"github.com/algorandfoundation/algorun-tui/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		return
	}
}
