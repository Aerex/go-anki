package card_types

import (
	"github.com/aerex/go-anki/pkg/anki"
	cmdList "github.com/aerex/go-anki/pkg/cmd/card-type/list"
	"github.com/spf13/cobra"
)

func NewCardTypeCmd(anki *anki.Anki) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "card-type <command>",
		Short: "Manage card types",
	}

	cmd.AddCommand(cmdList.NewListCmd(anki, nil))

	return cmd
}
