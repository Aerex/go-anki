package note_types

import (
	"github.com/aerex/go-anki/pkg/anki"
	cmdList "github.com/aerex/go-anki/pkg/cmd/note-type/list"
	"github.com/spf13/cobra"
)

func NewNoteTypeCmd(anki *anki.Anki) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "note-type <command>",
		Short: "Manage note types",
	}

	cmd.AddCommand(cmdList.NewListCmd(anki, nil))

	return cmd
}
