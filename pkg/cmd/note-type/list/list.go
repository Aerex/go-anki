package list

import (
	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/pkg/template"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	Query    string
	Template string
}

func NewListCmd(anki *anki.Anki, overrideF func(*anki.Anki) error) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List note types",
		RunE: func(cmd *cobra.Command, args []string) error {
			if overrideF != nil {
				return overrideF(anki)
			}
			return listCmd(anki, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Override template for output")

	return cmd
}

func listCmd(anki *anki.Anki, opts *ListOptions) error {
	tmpl := template.LIST_NOTE_TYPES
	if opts.Template != "" {
		tmpl = opts.Template
	}
	if err := anki.Templates.Load(tmpl); err != nil {
		return err
	}

	mdls, err := anki.Api.NoteTypes()
	if err != nil {
		return err
	}

	// TODO: 1. Retrieve models.tmpl
	// 2. Replace {{WORD}} with {{.WORD}}
	// 3. Convert html to something
	// 4. Convert inline css to something (use available css if available)
	data := struct {
		Data models.NoteTypes
	}{
		Data: mdls,
	}

	if err := anki.Templates.Execute(data, anki.IO); err != nil {
		return err
	}

	return nil
}
