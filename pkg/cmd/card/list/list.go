package list

import (
	"fmt"

	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/pkg/template"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	Query    string
	Template string
	Fields   []string
	Limit    int
}

func NewListCmd(anki *anki.Anki) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:   "list [-q, --query] [-t, --template] [-l, --limit]",
		Short: "List cards",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Limit < 0 {
				return fmt.Errorf("invalid limit: %v", opts.Limit)
			}
			return listCmd(anki, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Filter using expressions, see https://docs.ankiweb.net/searching.html")
	cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Format output using a Go template")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 30, "Maximum number of cards to return")

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

	cards, err := anki.Api.Cards(opts.Query, opts.Limit)
	if err != nil {
		return err
	}
	if len(cards) == 0 {
		return nil
	}

	var (
    QAs []models.CardQA
    cardTmpl models.CardTemplate
  )
  for idx, card := range cards {
      if card.Note.Model.Type == models.ClozeCardType {
        cardTmpl = *card.Note.Model.Templates[0]
      } else {
        for _, tmpl := range card.Note.Model.Templates {
          if tmpl.Ordinal == card.Ord {
            cardTmpl = *tmpl
            break
          }
        }
      }

    defer template.RecoverRender(cardTmpl, idx+1)
    QA, err := template.RenderCard(anki.Config, card, cardTmpl)
    if err != nil {
      return err
    }
    QAs = append(QAs, QA)
	}

	data := struct {
		Data []models.CardQA
	}{
		Data: QAs,
	}

	if err := anki.Templates.Execute(data, anki.IO); err != nil {
		return err
	}

	return nil
}
