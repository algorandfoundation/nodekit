package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/algorandfoundation/nodekit/ui/prompt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"golang.org/x/term"
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

// Confirm asks a yes/no question on the terminal. Anything but "y" or "yes",
// including just pressing Enter, is a no. Without a terminal to ask on, the
// answer is no.
func Confirm(question string) bool {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		log.Warn(fmt.Sprintf("%s (no terminal to ask on, answering no)", question))
		return false
	}
	fmt.Fprintf(os.Stderr, "%s [y/N] ", question)
	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}
