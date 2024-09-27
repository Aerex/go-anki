package create

import (
	"fmt"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/pkg/ui/prompt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type CreateCmd struct {
	Anki     *anki.Anki
	Type     string
	Quiet    bool
	File     string
	Tags     []string
	NoEditor bool
	Deck     string
	Fields   map[string]string
}

func NewCreateCmd(anki *anki.Anki) *cobra.Command {
	run := &CreateCmd{
		Anki: anki,
	}
	cmd := &cobra.Command{
		Use:                   "create [--type|-T TYPE] [--field|-f FIELD...] [--deck|-d DECK_NAME] [--file|-F FILE] [--tag|-t TAG...]",
		DisableFlagsInUseLine: true, // disables [flags] in usage text
		Short:                 "Create a card",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run.execute()
		},
	}

	cmd.Flags().StringVarP(&run.Type, "type", "T", "", "The note type")
	cmd.Flags().StringVarP(&run.File, "file", "F", "", "Read card content from file. Use \"-\" for stdin")
	cmd.Flags().StringVarP(&run.Deck, "deck", "d", "", "The name of the deck the new card will be added to")
	cmd.Flags().StringSliceVarP(&run.Tags, "tag", "t", []string{}, "List of tags on note")
	cmd.Flags().StringToStringVarP(&run.Fields, "field", "f", map[string]string{}, "Set the card fields")

	return cmd
}

func (cmd *CreateCmd) execute() (err error) {
	var (
		cardType string
		deckName string
	)
	note := models.Note{}
	prompt := prompt.NewSurveyPrompt(*cmd.Anki.Config)
	if cmd.Type != "" {
		cardType = cmd.Type
	} else {
		noteTypes, err := cmd.Anki.API.NoteTypes()
		if err != nil {
			return err
		}
		var noteTypeNames []string
		for _, noteType := range noteTypes {
			noteTypeNames = append(noteTypeNames, noteType.Name)
		}
		cardType, err = prompt.Choose("Type:", noteTypeNames, "Basic")
		if err != nil {
			return err
		}
	}
	if cmd.Deck != "" {
		deckName = cmd.Deck
	} else {
		decks, err := cmd.Anki.API.Decks("")
		if err != nil {
			return err
		}
		deckNames := make([]string, len(decks))
		for idx, deck := range decks {
			deckNames[idx] = deck.Name
		}
		deckName, err = prompt.Choose("Deck:", deckNames, "Default")
		if err != nil {
			return err
		}
	}

	noteType, err := cmd.Anki.API.NoteType(cardType)
	if err != nil {
		return err
	}
	noteTypeMap := make(map[string]string)
	if len(cmd.Fields) > 0 {
		for _, f := range noteType.Fields {
			noteTypeMap[f.Name] = f.Name
		}
		invalidFields := []string{}
		fields := []string{}
		for _, value := range cmd.Fields {
			if _, ok := noteTypeMap[value]; ok {
				fields = append(fields, value)
			} else {
				invalidFields = append(invalidFields, value)
			}
		}

		if len(invalidFields) > 0 {
			return fmt.Errorf("fields \"%s\" are not defined in %s", strings.Join(invalidFields, ", "), noteType.Name)
		}
		note.Fields = fields
	} else {
		fieldPrompts := make([]*survey.Question, len(noteType.Fields))
		fieldResponses := make(map[string]interface{}, len(noteType.Fields))
		for idx, field := range noteType.Fields {
			fieldPrompts[idx] = &survey.Question{
				Prompt: &survey.Input{Message: field.Name},
				Name:   field.Name,
			}
		}
		err = survey.Ask(fieldPrompts, &fieldResponses)
		if err != nil {
			return err
		}
		note.Fields = make([]string, len(noteType.Fields))
		for idx, f := range noteType.Fields {
			if field, ok := fieldResponses[f.Name]; ok {
				note.Fields[idx] = field.(string)
			}
		}
	}
	if len(cmd.Tags) > 0 {
		note.StringTags = strings.Join(cmd.Tags, ",")
	} else {
		includeTags, err := prompt.Confirm("Add tags?")
		if err != nil {
			return err
		}
		if includeTags {
			tags, err := cmd.Anki.API.Tags()
			if err != nil {
				return err
			}
			selectedTags, err := prompt.Select("Tags:", tags, tags[0])
			if err != nil {
				return err
			}
			note.StringTags = strings.Join(selectedTags, ",")
		}
	}
	_, err = cmd.Anki.API.CreateCard(note, noteType, deckName)
	if err != nil {
		log.Logger.Error().Err(err)
		return err
	}
	return nil
}
