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

func NewListCmd(anki *anki.Anki, cb func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:   "list <options>",
		Short: "List decks",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cb != nil {
				return cb(opts)
			}
			return listCmd(anki, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Filter using expressions, see https://docs.ankiweb.net/searching.html")
	cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Override template for output")

	return cmd
}

func listCmd(anki *anki.Anki, opts *ListOptions) error {
	tmpl := template.LIST_DECK
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

	decks, err := anki.API.Decks(query)
	if err != nil {
		return err
	}

	statsByDeck, err := anki.API.DeckStudyStats()
	if err != nil {
		return err
	}
	stats := []models.DeckStudyStats{}
	for _, deck := range decks {
		stats = append(stats, statsByDeck[deck.ID])
	}

	data := struct {
		Data []*models.Deck
		Meta []models.DeckStudyStats
	}{
		Data: decks,
		Meta: stats,
	}

	if err := anki.Templates.Execute(data, anki.IO); err != nil {
		return err
	}

	return nil
}
