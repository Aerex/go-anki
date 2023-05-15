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
	Name     string
}

func NewListCmd(anki *anki.Anki, overrideF func(*anki.Anki) error) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List card types",
		RunE: func(cmd *cobra.Command, args []string) error {
			if overrideF != nil {
				return overrideF(anki)
			}
			return listCmd(anki, opts)
		},
	}

	// TODO: No way to query note types since data is a json blob
	// Potential solutions could be to index note types into sepeate data base for querying
	//cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Filter using expressions, see https://docs.ankiweb.net/searching.html")
	cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Override template for output")
	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name of card type")

	return cmd
}

func listCmd(anki *anki.Anki, opts *ListOptions) error {
	tmpl := template.LIST_CARD_TYPES
	if opts.Template != "" {
		tmpl = opts.Template
	}
	if err := anki.Templates.Load(tmpl); err != nil {
		return err
	}

	var name string
	if opts.Name != "" {
		name = opts.Name
	}

	mdls, err := anki.Api.GetModels(name)
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
