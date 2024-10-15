package main

import (
	"github.com/algorandfoundation/hack-tui/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		return
	}
}
