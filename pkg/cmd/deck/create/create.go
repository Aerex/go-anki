package create

import (
	"bytes"

	"github.com/aerex/go-anki/pkg/anki"
	"github.com/spf13/cobra"
)

type CreateDeckOptions struct {
	Type string
}

func NewCreateCmd(anki *anki.Anki, cb func(*anki.Anki) error) *cobra.Command {
	opts := &CreateDeckOptions{}

	cmd := &cobra.Command{
		Use:          "create [deck_name]",
		Short:        "Create a deck",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cb != nil {
				return cb(anki)
			}
			return createCmd(anki, args, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Type, "type", "t", "Basic", "The type of deck")

	return cmd
}

func createCmd(anki *anki.Anki, args []string, opts *CreateDeckOptions) error {
	deck, err := anki.Api.CreateDeck(args[0], opts.Type)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	buffer.WriteString("Created deck " + deck.Name)
	buffer.WriteTo(anki.IO.Output)

	return nil
}
