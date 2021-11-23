package card

import (
	"github.com/aerex/anki-cli/pkg/anki"
	cmdList "github.com/aerex/anki-cli/pkg/cmd/card/list"
	"github.com/spf13/cobra"
)

func NewCardCmd(anki *anki.Anki) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "card <command>",
		Short: "Manage cards",
	}

	cmd.AddCommand(cmdList.NewListCmd(anki, nil))

	return cmd
}
