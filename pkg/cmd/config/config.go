package config

import (
	"github.com/aerex/anki-cli/pkg/anki"
	cmdList "github.com/aerex/anki-cli/pkg/cmd/decks/list"
	"github.com/spf13/cobra"
)


func NewCmdTemplates(anki *anki.Anki) *cobra.Command {
  cmd := &cobra.Command{
    Use: "templates <command>",
    Short: "Manage templates",
  }

  cmd.AddCommand(cmdList.NewListCmd(anki, nil))

  return cmd
}
