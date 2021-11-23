package list

import (
	"github.com/aerex/anki-cli/pkg/anki"
	"github.com/aerex/anki-cli/pkg/models"
	"github.com/aerex/anki-cli/pkg/template"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	Query    string
	Template string
}

func NewListCmd(anki *anki.Anki, overrideF func(*anki.Anki) error) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:   "list [-q quiet] [-t template]",
		Short: "List cards",
		RunE: func(cmd *cobra.Command, args []string) error {
			if overrideF != nil {
				return overrideF(anki)
			}
			return listCmd(anki, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Filter using expressions, see https://docs.ankiweb.net/searching.html")
	cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Override template for output")

	return cmd
}

func listCmd(anki *anki.Anki, opts *ListOptions) error {
	tmpl := template.CARD_LIST
	if opts.Template != "" {
		tmpl = opts.Template
	}
	if err := anki.Templates.Load(tmpl); err != nil {
		return err
	}

	var query string
	if opts.Query != "" {
		query = opts.Query
	}

	cards, err := anki.Api.GetCards(query)
	if err != nil {
		return err
	}

	data := struct {
		Data []models.Card
	}{
		Data: cards,
	}

	if err := anki.Templates.Execute(data, anki.IO); err != nil {
		return err
	}

	return nil
}
