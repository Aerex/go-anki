package root

import (
	"github.com/aerex/go-anki/pkg/anki"
	cardCommand "github.com/aerex/go-anki/pkg/cmd/card"
	deckCommand "github.com/aerex/go-anki/pkg/cmd/deck"
	deckConfigCommand "github.com/aerex/go-anki/pkg/cmd/deck-config"
	noteTypeCommand "github.com/aerex/go-anki/pkg/cmd/note-type"
	studyCommand "github.com/aerex/go-anki/pkg/cmd/study"
	"github.com/spf13/cobra"
)

func NewRootCmd(anki *anki.Anki) *cobra.Command {
	root := &cobra.Command{
		Use:   "anki",
		Short: "Interact with Anki from the terminal",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Usage()
			}
		},
	}

	root.SetOut(anki.IO.Output)
	root.SetErr(anki.IO.Error)

	root.AddCommand(deckCommand.NewCmdDeck(anki))
	root.AddCommand(cardCommand.NewCardCmd(anki))
	root.AddCommand(deckConfigCommand.NewDeckConfigsCmd(anki, nil))
	root.AddCommand(noteTypeCommand.NewNoteTypeCmd(anki))
	root.AddCommand(studyCommand.NewStudyCmd(anki))

	return root
}
