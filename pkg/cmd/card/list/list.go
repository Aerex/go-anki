package list

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"runtime/debug"

	"strings"

	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/pkg/template"
	"github.com/microcosm-cc/bluemonday"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
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

	var QAs []models.CardQA

	// TODO: externalize into  method for produing questio and answer
	var errCardNums []int
	var jsonStr strings.Builder
	p := bluemonday.StrictPolicy()
	var fieldStruct dynamicstruct.Builder
	var fieldMap map[string]string
	for idx_c, card := range cards {
		err := func() error {
			jsonStr.WriteString("{")
			fieldStruct = dynamicstruct.NewStruct()
			fieldMap = make(map[string]string)
			for idx, field := range card.Note.Model.Fields {
				name := field.Name
				value := card.Note.Fields[field.Ordinal]
				fieldName := fmt.Sprintf("Field_%d", idx)
				fieldMap[name] = fieldName
				fieldStruct.AddField(fieldName, "", `json:"`+fieldName+`"`)
				fmt.Fprintf(&jsonStr, `"%s": "%s"`, fieldName, p.Sanitize(strings.ReplaceAll(value, "\"", "\\\"")))
				if idx < len(card.Note.Fields)-1 {
					jsonStr.WriteString(",")
				}
			}
			jsonStr.WriteString("}")
			instance := fieldStruct.Build().New()
			err = json.Unmarshal([]byte(jsonStr.String()), &instance)
			if err != nil {
				return err
			}
			readStruct := dynamicstruct.NewReader(instance)
			jsonStr.Reset()

			var qfmt string
			var afmt string
			var err error
			var qb bytes.Buffer
			var ab bytes.Buffer
			var cardTmpl models.CardTemplate
			// TODO: May need to sort this instead of iterating
			// NOTE: Could be slow if we have many templates per card
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

			qfmt, err = template.ParseCardTemplate(template.TemplateParseOptions{
				CardTemplate:     cardTmpl,
				CardTemplateName: cardTmpl.Name,
				Card:             card,
				IsAnswer:         false,
				ReadStruct:       readStruct,
				FieldMap:         fieldMap,
			})

			t, err := template.LoadString(qfmt, anki.Config, template.RENDER_LIST)
			if err != nil {
				return err
			}
			// NOTE: Need to use an inline method for defer here to allow
			// capture recover/panic errors
			defer func(tmpl models.CardTemplate, cardNum int) {
				if recoverErr := recover(); recoverErr != nil {
					re := regexp.MustCompile(`/text/template/exec`)
					if templateExecErr := re.Match(debug.Stack()); templateExecErr {
						errCardNums = append(errCardNums, cardNum)
						fmt.Printf("\nERROR: Could not generate card %d due to an issue with one its templates\n", cardNum)
						fmt.Println("\nQuestionTemplate: ", strings.ReplaceAll(html.UnescapeString(p.Sanitize(tmpl.QuestionFormat)), "\n", ""))
						fmt.Println("\nAnswerTemplate: ", strings.ReplaceAll(html.UnescapeString(p.Sanitize(tmpl.AnswerFormat)), "\n", ""))
					} else {
						debug.PrintStack()
					}
				}
			}(cardTmpl, (idx_c + 1))

			t.Execute(&qb, instance)

			afmt, err = template.ParseCardTemplate(template.TemplateParseOptions{
				CardTemplate:     cardTmpl,
				CardTemplateName: cardTmpl.Name,
				Card:             card,
				IsAnswer:         true,
				ReadStruct:       readStruct,
				FieldMap:         fieldMap,
			})

			// When we see  <hr id=answer> it is assumed that the content afterwards
			// is the answer so use that for answer format
			if strings.Contains(afmt, `<hr id=answer>`) {
				parts := strings.SplitAfter(afmt, "<hr id=answer>")
				if len(parts) > 1 {
					afmt = parts[1]
				}
			}

			t, err = template.LoadString(afmt, anki.Config, template.RENDER_LIST)
			if err != nil {
				return err
			}

			t.Execute(&ab, instance)

			QAs = append(QAs, models.CardQA{
				CardType: card.Note.Model.Name,
				Deck:     card.Deck.Name,
				Due:      card.Due,
				Question: strings.TrimSpace(html.UnescapeString(p.Sanitize(qb.String()))),
				Answer:   strings.TrimSpace(html.UnescapeString(p.Sanitize(ab.String()))),
			})
			return nil
		}()
		if err != nil {
			return err
		}
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
