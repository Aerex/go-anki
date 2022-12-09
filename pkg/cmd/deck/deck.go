package deck

import (
	"github.com/aerex/go-anki/pkg/anki"
	cmdCreate "github.com/aerex/go-anki/pkg/cmd/deck/create"
	cmdList "github.com/aerex/go-anki/pkg/cmd/deck/list"
	cmdRename "github.com/aerex/go-anki/pkg/cmd/deck/rename"
	"github.com/spf13/cobra"
)

func NewCmdDeck(anki *anki.Anki) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deck <command>",
		Short: "Manage decks",
	}

	cmd.AddCommand(cmdList.NewListCmd(anki, nil))
	cmd.AddCommand(cmdRename.NewRenameCmd(anki, nil))
	cmd.AddCommand(cmdCreate.NewCreateCmd(anki, nil))

	return cmd
}
