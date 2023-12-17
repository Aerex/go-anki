package rename

import (
	"bytes"

	"github.com/aerex/go-anki/pkg/anki"
	"github.com/spf13/cobra"
)

type RenameOptions struct {
	Quiet bool
}

func NewRenameCmd(anki *anki.Anki, cb func(*anki.Anki) error) *cobra.Command {
	opts := &RenameOptions{}

	cmd := &cobra.Command{
		Use:          "rename <name> <new_name> <options>",
		Short:        "Rename deck",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cb != nil {
				return cb(anki)
			}
			return renameCmd(anki, opts, args)
		},
	}

	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "Supress output from terminal and logs")

	return cmd
}

func renameCmd(anki *anki.Anki, opts *RenameOptions, args []string) error {

	err := anki.Api.RenameDeck(args[0], args[1])
	if err != nil {
		return err
	}

	if !opts.Quiet {
		var buffer bytes.Buffer
		buffer.WriteString("Renamed deck to " + args[1])
		buffer.WriteTo(anki.IO.Output)
	}
	return nil
}
