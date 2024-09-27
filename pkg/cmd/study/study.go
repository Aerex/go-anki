package study

import (
	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/pkg/template"
	"github.com/spf13/cobra"
)

func NewStudyCmd(anki *anki.Anki) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "study <deck_name>",
		Short: "Study a deck",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			return run(anki, args)
		},
	}
	return cmd
}

func run(anki *anki.Anki, args []string) error {
	deckName := args[0]
	var studyDeck models.Deck
	cards, err := anki.API.Cards("deck:"+args[0], -1)
	if err != nil {
		anki.IO.Log.Err(err).Msgf("failed to find cards for deck %s", deckName)
		return err
	}
	if len(cards) == 0 {
		anki.IO.Log.Info().Msgf("no cards to study for deck %s", deckName)
		return nil
	} else {
		// TODO: add the ability to select from decks that share the same name
		studyDeck = cards[0].Deck
	}

	if err != nil {
		anki.IO.Log.Err(err).Msgf("failed to list cards for deck %s", deckName)
		return err
	}

	cardQAs := make([]*models.CardQA, len(cards))
	var cardTmpl models.CardTemplate

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
		cardQA, err := template.RenderCard(anki.Config, card, cardTmpl)
		if err != nil {
			return err
		}
		cardQAs[idx] = &cardQA
	}

	stats, err := anki.API.DeckStudyStats()
	if err != nil {
		return err
	}

	if err := anki.API.StudyReview(anki.Log, studyDeck.Name, cardQAs, stats[studyDeck.ID]); err != nil {
		return err
	}

	// Create terminal app
	// For each card in deck
	// Fill terminal app with
	// -- card template
	// -- show answer button

	return nil
}
