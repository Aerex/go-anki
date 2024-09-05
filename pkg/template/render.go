package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/aerex/go-anki/internal/config"
	"github.com/microcosm-cc/bluemonday"
	dynamicstruct "github.com/ompluscator/dynamic-struct"

	"github.com/aerex/go-anki/pkg/models"
)

var ParsePolicy = bluemonday.StrictPolicy()

func RenderCard(config *config.Config, card models.Card, cardTmpl models.CardTemplate) (models.CardQA, error) {
	readStruct, fieldMap, cardErr := generateCardStruct(card)
	if cardErr != nil {
		return models.CardQA{}, cardErr
	}
	tmplParseOpts := TemplateParseOptions{
		CardTemplate: cardTmpl,
		Card:         card,
		ReadStruct:   readStruct,
		FieldMap:     fieldMap,
	}

	var questionBuffer bytes.Buffer
	tmplParseOpts.IsAnswer = false
	err := templateFromCard(config, tmplParseOpts, &questionBuffer)
	if err != nil {
		return models.CardQA{}, err
	}

	var (
		answerQuestionBuffer bytes.Buffer
		answerOnly           string
	)
	tmplParseOpts.IsAnswer = true
	err = templateFromCard(config, tmplParseOpts, &answerQuestionBuffer)
	if err != nil {
		return models.CardQA{}, err
	}

	// When we see  <hr id=answer> it is assumed that the content afterwards
	// is the answer so use that for answer format
	if strings.Contains(answerQuestionBuffer.String(), "<hr id=answer>") {
		parts := strings.SplitAfter(answerQuestionBuffer.String(), "<hr id=answer>")
		if len(parts) > 1 {
			answerOnly = parts[1]
		}
	}

	return models.CardQA{
		Card:            card,
		Question:        strings.TrimSpace(html.UnescapeString(ParsePolicy.Sanitize(questionBuffer.String()))),
		QuestionBrowser: questionBuffer.String(),
		Answer:          strings.TrimSpace(html.UnescapeString(ParsePolicy.Sanitize(answerOnly))),
		AnswerBrowser:   answerQuestionBuffer.String(),
	}, nil
}

func generateCardStruct(card models.Card) (dynamicstruct.Reader, map[string]string, error) {
	jsonStr := strings.Builder{}
	jsonStr.WriteString("{")
	fieldStruct := dynamicstruct.NewStruct()
	fieldMap := make(map[string]string)

	for idx, field := range card.Note.Model.Fields {
		name := field.Name
		value := card.Note.Fields[field.Ordinal]
		fieldName := fmt.Sprintf("Field_%d", idx)
		fieldMap[name] = fieldName
		fieldStruct.AddField(fieldName, "", `json:"`+fieldName+`"`)
		fmt.Fprintf(&jsonStr, `"%s": "%s"`, fieldName, html.UnescapeString(ParsePolicy.Sanitize(strings.ReplaceAll(value, "\"", "\\\""))))
		if idx < len(card.Note.Fields)-1 {
			jsonStr.WriteString(",")
		}
	}
	jsonStr.WriteString("}")
	instance := fieldStruct.Build().New()
	if err := json.Unmarshal([]byte(jsonStr.String()), &instance); err != nil {
		return nil, nil, err
	}
	return dynamicstruct.NewReader(instance), fieldMap, nil
}
func templateFromCard(config *config.Config, opts TemplateParseOptions, tmplOut *bytes.Buffer) error {
	tmplFmt, err := ParseCardTemplate(opts)
	if err != nil {
		return err
	}

	t, err := LoadString(tmplFmt, config, RENDER_LIST)
	if err != nil {
		return err
	}
	if err := t.Execute(tmplOut, opts.ReadStruct.GetValue()); err != nil {
		return fmt.Errorf("Could not generate card from template: %s", tmplFmt)
	}
	return nil
}

func RecoverRender(tmpl models.CardTemplate, cardNum int) {
	if recoverErr := recover(); recoverErr != nil {
		re := regexp.MustCompile(`/text/template/exec`)
		if templateExecErr := re.Match(debug.Stack()); templateExecErr {
			fmt.Printf("\nFATAL: Could not generate card %d due to an issue with one its templates\n", cardNum)
			fmt.Println("\nQuestionTemplate: ", strings.ReplaceAll(html.UnescapeString(ParsePolicy.Sanitize(tmpl.QuestionFormat)), "\n", ""))
			fmt.Println("\nAnswerTemplate: ", strings.ReplaceAll(html.UnescapeString(ParsePolicy.Sanitize(tmpl.AnswerFormat)), "\n", ""))
		} else {
			debug.PrintStack()
		}
	}
}
