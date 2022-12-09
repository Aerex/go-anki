package create

import (
	"bytes"

	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/editor"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/pkg/template"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
)

var logger = logging.MustGetLogger("ankicli")

type CreateOptions struct {
	Type     string
	Template string
	NoEditor bool
	Deck     string
	Fields   map[string]string
}

func NewCreateCmd(anki *anki.Anki, overrideF func(*anki.Anki) error) *cobra.Command {
	opts := &CreateOptions{}

	cmd := &cobra.Command{
		Use:                   "create (-t type|Basic) [-f field1 -f field2...] [-d deck|Default]",
		DisableFlagsInUseLine: true, // disables [flags] in usage text
		Short:                 "Create a card",
		RunE: func(cmd *cobra.Command, args []string) error {
			if overrideF != nil {
				return overrideF(anki)
			}
			if len(args) > 0 && args[0] == "help" {
				return cmd.Usage()
			}
			return createCmd(anki, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Type, "type", "t", "Basic", "The note type")
	cmd.Flags().StringVarP(&opts.Deck, "deck", "d", "Default", "The name of the deck the new card will be added to")
	// FIXME: Need to use another character option for template
	cmd.Flags().StringVar(&opts.Template, "template", "", "Override template for create a card")
	cmd.Flags().BoolVar(&opts.NoEditor, "no-edit", false, "Disable using the editor")
	cmd.Flags().StringToStringVarP(&opts.Fields, "field", "f", map[string]string{}, "Set the card fields")

	return cmd
}

func createCmd(anki *anki.Anki, opts *CreateOptions) error {
	tmpl := template.CREATE_CARD
	cardType := "Basic"
	deckName := "Default"
	note := models.Note{}
	if opts.Type != "" {
		cardType = opts.Type
	}

	if opts.Deck != "" {
		deckName = opts.Deck
	}

	if len(opts.Fields) > 0 {
		fields := []string{}
		for _, value := range opts.Fields {
			fields = append(fields, value)
		}
		note.Fields = fields
	}

	if opts.Template != "" {
		tmpl = opts.Template
	}
	if err := anki.Templates.Load(tmpl); err != nil {
		return err
	}

	noteType, err := anki.Api.GetNoteType(cardType)

	if err != nil {
		logger.Error(err)
		return err
	}

	if !opts.NoEditor {
		edit := editor.NewModelEditor(anki, &note, &noteType)
		edit.Create()
		defer edit.Remove()
		for {
			err, changed := edit.EditNote()
			if !changed {
				return nil
			}

			if err == nil {
				break
			}
			retry := edit.ConfirmUserError()
			if !retry {
				logger.Error(err)
				return err
			}
		}
	}

	// TODO: Figure out what to do with output card
	_, err = anki.Api.CreateCard(note, noteType, deckName)
	if err != nil {
		logger.Error(err)
		return err
	}

	var buffer bytes.Buffer
	buffer.WriteString("Created new card")
	buffer.WriteTo(anki.IO.Output)

	return nil
}
