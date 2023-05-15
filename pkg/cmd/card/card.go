package card

import (
	"github.com/aerex/go-anki/pkg/anki"
	cmdCreate "github.com/aerex/go-anki/pkg/cmd/card/create"
	cmdList "github.com/aerex/go-anki/pkg/cmd/card/list"
	"github.com/spf13/cobra"
)

func NewCardCmd(anki *anki.Anki) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "card <command>",
		Short: "Manage cards",
	}

	cmd.AddCommand(cmdList.NewListCmd(anki))
	cmd.AddCommand(cmdCreate.NewCreateCmd(anki))

	return cmd
}
