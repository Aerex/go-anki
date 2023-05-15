package template

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aerex/go-anki/pkg/models"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
	"github.com/stretchr/testify/assert"
)

type templateFmts struct {
	qfmt string
	afmt string
}

type cardData struct {
	Field_0 string
	Field_1 string
	Cloze_1 string
}

func TestParseCardTemplate(t *testing.T) {
	card := models.Card{
		Deck: models.Deck{
			Name: "English::Vocab",
		},
		Note: models.Note{
			Model: models.NoteType{
				Name: "Test",
			},
			StringTags: "tag_a,tag_b",
		},
	}
	fieldMap := map[string]string{
		"Front": "Field_0",
		"Back":  "Field_1",
		"Cloze": "Cloze_1",
	}
	cardDataInst := dynamicstruct.ExtendStruct(cardData{}).Build().New()
	data := []byte(`{"Field_0": "This is the question", "Field_1": "This is the answer", "Cloze_1": "{{c1::hidden text}}"}`)
	err := json.Unmarshal(data, &cardDataInst)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal dynamic struct %#v due to: %v", cardDataInst, err.Error()))
	}

	readStruct := dynamicstruct.NewReader(cardDataInst)

	tests := []struct {
		name string
		// { "qfmt", "afmt" }
		goTemplates   []string
		ankiTemplates []string
		card          models.Card
		isError       bool
	}{
		{
			name:          "Template using FrontSide special field on list card",
			goTemplates:   []string{"{{.Field_0}}", "{{.Field_0}} \n\n<hr id=answer>\n\n{{.Field_1}}"},
			ankiTemplates: []string{"{{Front}}", "{{FrontSide}} \n\n<hr id=answer>\n\n{{Back}}"},
		},
		{
			name:          "Template using Tags special field on list card",
			goTemplates:   []string{"tag_a,tag_b {{.Field_0}}", "<hr id=answer>\n\n{{.Field_1}}"},
			ankiTemplates: []string{"{{Tags}} {{Front}}", "<hr id=answer>\n\n{{Back}}"},
		},
		{
			name:          "Template using Type special field on list card",
			goTemplates:   []string{"Test ", "<hr id=answer>\n\n{{.Field_1}}"},
			ankiTemplates: []string{"{{Type}}", "<hr id=answer>\n\n{{Back}}"},
		},
		{
			name:          "Template using Subdeck special field on list card",
			goTemplates:   []string{"Vocab ", "<hr id=answer>\n\n{{.Field_1}}"},
			ankiTemplates: []string{"{{Subdeck}}", "<hr id=answer>\n\n{{Back}}"},
		},
		{
			name:          "Template using Card special field on list card",
			goTemplates:   []string{"Basic ", "<hr id=answer>\n\n{{.Field_1}}"},
			ankiTemplates: []string{"{{Card}}", "<hr id=answer>\n\n{{Back}}"},
		},
		{
			name:          "Template using Hint field on list card",
			goTemplates:   []string{"{{hint .Field_0}}", "<hr id=answer>\n\n{{hint .Field_1}}"},
			ankiTemplates: []string{"{{hint:Front}}", "<hr id=answer>\n\n{{hint:Back}}"},
		},
		{
			name:          "Template using filter with arguments",
			goTemplates:   []string{"{{tts .Field_0 \"ja_JP\" \"voices=Apple_Otoya,Microsoft_Haruka\"}}", "{{tts .Field_1 \"fr_FR\" \"speed=0.8\"}}"},
			ankiTemplates: []string{"{{tts ja_JP voices=Apple_Otoya,Microsoft_Haruka:Front}}", "{{tts fr_FR speed=0.8:Back}}"},
		},
		{
			name:          "Template using cloze",
			goTemplates:   []string{"{{ cloze \"1\" \"1\" \"hidden text\" \"0\" }}", "{{ cloze \"1\" \"1\" \"hidden text\" \"1\" }}"},
			ankiTemplates: []string{"{{cloze:Cloze}}", "{{cloze:Cloze}}"},
		},
	}

	for _, tt := range tests {
		assert.Equal(t, len(tt.ankiTemplates), len(tt.goTemplates))

		opts := TemplateParseOptions{
			CardTemplateName: "Basic",
			CardTemplate: models.CardTemplate{
				Name:           "Basic",
				QuestionFormat: tt.ankiTemplates[0],
				AnswerFormat:   tt.ankiTemplates[1],
			},
			Card:       card,
			ReadStruct: readStruct,
			FieldMap:   fieldMap,
		}

		for idx := range tt.ankiTemplates {
			if idx == 0 {
				opts.IsAnswer = false
			} else {
				opts.IsAnswer = true
			}
			parsedGoTmpl, err := ParseCardTemplate(opts)

			if tt.isError {
				assert.Error(t, err)
				assert.Empty(t, parsedGoTmpl)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.goTemplates[idx], parsedGoTmpl)
			}
		}
	}
}
