package create

import (
	"bytes"

	"github.com/aerex/anki-cli/pkg/anki"
	"github.com/spf13/cobra"
)


func NewCreateCmd(anki *anki.Anki, overrideF func(*anki.Anki) error) *cobra.Command {

  cmd := &cobra.Command {
    Use: "create [deck_name]",
    Short: "Create a deck",
    Args: cobra.ExactArgs(1),
    SilenceUsage: true,
    RunE: func(cmd *cobra.Command, args []string) error {
      if overrideF != nil {
        return overrideF(anki)
      }
      return createCmd(anki, args)
    },
  }

  return cmd
}

func createCmd(anki *anki.Anki, args []string) error {

  deck, err := anki.Api.CreateDeck(args[0])
  if err != nil {
    return err
  }

  var buffer bytes.Buffer
  buffer.WriteString("Created deck " + deck.Name)
  buffer.WriteTo(anki.IO.Output)

  return nil
}
