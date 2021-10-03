package decks

import (
	"github.com/aerex/anki-cli/pkg/anki"
	cmdList "github.com/aerex/anki-cli/pkg/cmd/deck/list"
	"github.com/spf13/cobra"
)


func NewCmdDeck(anki *anki.Anki) *cobra.Command {
  cmd := &cobra.Command{
    Use: "deck <command>",
    Short: "Manage decks",
  }

  cmd.AddCommand(cmdList.NewListCmd(anki, nil))

  return cmd
}
