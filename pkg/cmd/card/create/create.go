package create

import (
	"bytes"
	"io/ioutil"
	"strings"

	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/pkg/template"
	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var logger = logging.MustGetLogger("ankicli")

type CreateOptions struct {
	Type     string
	Quiet    bool
	File     string
	Template string
	NoEditor bool
	Deck     string
	Fields   map[string]string
}

func NewCreateCmd(anki *anki.Anki) *cobra.Command {
	opts := &CreateOptions{}

	cmd := &cobra.Command{
		Use:                   "create (-t type|Basic) [-f field1 -f field2...] [-d deck|Default]",
		DisableFlagsInUseLine: true, // disables [flags] in usage text
		Short:                 "Create a card",
		RunE: func(cmd *cobra.Command, args []string) error {
			return createCmd(anki, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Type, "type", "t", "Basic", "The note type")
	cmd.Flags().StringVarP(&opts.File, "file", "F", "", "Read card content from file. Use \"-\" for stdin")
	cmd.Flags().BoolVarP(&opts.Quiet, "quiet", "q", false, "Silence")
	cmd.Flags().StringVarP(&opts.Deck, "deck", "d", "Default", "The name of the deck the new card will be added to")
	cmd.Flags().StringVar(&opts.Template, "template", "", "Override template for create a card")
	cmd.Flags().BoolVar(&opts.NoEditor, "no-edit", false, "Disable using the editor")
	cmd.Flags().StringToStringVarP(&opts.Fields, "field", "f", map[string]string{}, "Set the card fields")

	return cmd
}

func createCmd(anki *anki.Anki, opts *CreateOptions) error {
	var (
		cardType string
		deckName string
	)
	tmpl := template.CREATE_CARD
	note := models.Note{}
	createNote := models.CreateNote{}
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

	var (
		data    []byte
		changed bool
	)
	if opts.File == "" {
		if err := anki.Editor.Create(); err != nil {
			return err
		}
		defer anki.Editor.Remove()
		for {
			err, data, changed = anki.Editor.Edit(noteType)
			if !changed {
				return nil
			}

			if err == nil {
				break
			}
			retry := anki.Editor.ConfirmUserError()
			if !retry {
				logger.Error(err)
				return err
			}
		}
	} else {
		data, err = ioutil.ReadFile(opts.File)
		if err != nil {
			return err
		}
	}

	if err := yaml.Unmarshal(data, &createNote); err != nil {
		return err
	}

	// TODO: May need to move this to a method
	note.Model = noteType
	for _, fld := range createNote.Fields {
		note.Fields = append(note.Fields, fld)
	}
	note.StringTags = strings.Join(createNote.Tags, ",")

	// TODO: Figure out what to do with output card,
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
