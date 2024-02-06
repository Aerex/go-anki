package template

import (
	"fmt"
	"strings"

	"github.com/aerex/go-anki/pkg/models"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

type TemplateParseOptions struct {
	IsAnswer         bool
	CardTemplate     models.CardTemplate
  CardTemplateName string
	Card             models.Card
	ReadStruct       dynamicstruct.Reader
	FieldMap         map[string]string

	inTag       bool
	isFrontSide bool
	fieldTag    string
}

func ParseCardTemplate(opts TemplateParseOptions) (string, error) {
	var output strings.Builder
	frOpts := FieldReplacmentOptions{
		card:        opts.Card,
		isFrontSide: opts.isFrontSide,
		isAnswer:    opts.IsAnswer,
		readStruct:  opts.ReadStruct,
		cardTmpl:    opts.CardTemplate,
		fieldMap:    opts.FieldMap,
	}
	var tmplFmt string
	if opts.IsAnswer {
		tmplFmt = opts.CardTemplate.AnswerFormat
	} else {
		tmplFmt = opts.CardTemplate.QuestionFormat
	}
	token := NewTokenizer(tmplFmt)
	for token.HasNext() {
		tt := token.Next()
		switch tt {
		case EndOfBuffer:
			// Exceeded the buffer
			// so return what we have in output
			return output.String(), nil
		case TextToken:
			if opts.inTag {
				if _, err := ParseCardTemplate(opts); err != nil {
					return "", err
				}
			}
			output.WriteString(token.Text())
		case FieldReplacementToken:
			frOpts.tmplFmt = token.Text()
			fieldReplTmpl, err := FieldReplacement(frOpts)
			if err != nil {
				return "", err
			}
			output.WriteString(fieldReplTmpl)
		case OpenConditionalToken:
			if field := opts.ReadStruct.GetField(token.Text()); field != nil {
				opts.inTag = true
				opts.fieldTag = token.Text()
				condTmpl, err := ParseCardTemplate(opts)
				if err != nil {
					return "", err
				}
				output.WriteString(condTmpl)
			}
		case CloseConditionalToken:
			if !opts.inTag {
				return "", fmt.Errorf("missing closing \"{{\" in template %s", tmplFmt)
			}
			closeTagName := token.Text()
			if closeTagName != opts.fieldTag {
				return "", fmt.Errorf("found \"{{\\%s}}\" conditional end tag; expected \"{{\\%s}}\"", closeTagName, opts.fieldTag)
			}
			opts.inTag = false
			opts.fieldTag = ""
		}
	}

	return output.String(), nil
}

