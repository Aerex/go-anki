package root

import (
	"github.com/aerex/anki-cli/pkg/anki"
	deckCommand "github.com/aerex/anki-cli/pkg/cmd/deck"
	"github.com/spf13/cobra"
)

func NewRootCmd(anki *anki.Anki) *cobra.Command { 
  root := &cobra.Command {
    Use: "anki",
    Short: "Interact with Anki from the terminal",
    Run: func(cmd *cobra.Command, args []string) {

      },
  }

  root.SetOut(anki.IO.Output)
  root.SetErr(anki.IO.Error)

  root.AddCommand(deckCommand.NewCmdDeck(anki))

  return root
}
