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
		Use:          "create <name>",
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
	return cmd
}

func createCmd(anki *anki.Anki, args []string, opts *CreateDeckOptions) error {
	if err := anki.API.CreateDeck(args[0]); err != nil {
		return err
	}

	var buffer bytes.Buffer
	buffer.WriteString("Created deck " + args[0])
	buffer.WriteTo(anki.IO.Output)

	return nil
}
