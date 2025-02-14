package utils

import (
	"errors"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/spf13/cobra"
)

func IsSudoCmd(cmd *cobra.Command, args []string) error {
	if !system.IsSudo() {
		return errors.New(explanations.NotSuperUserErrorMsg)
	}
	return nil
}
