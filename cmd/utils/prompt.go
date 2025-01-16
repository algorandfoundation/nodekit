package utils

import (
	"github.com/algorandfoundation/nodekit/ui/prompt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

func Prompt(message string) bool {
	promptUi := prompt.New(message)

	p := tea.NewProgram(promptUi)
	answer := false
	go func() {
		for {
			response := <-promptUi.Outside
			answer = response.(bool)
		}
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	return answer
}
