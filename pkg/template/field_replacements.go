package template

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/internal/utils"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/third_party/kakasi"
	"github.com/mgutz/ansi"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

var (
	REGEX_MATCH_TMPL_ARG_GROUPS = regexp.MustCompile(`([\w=,\.]+)[^\\h]`)
	REGEX_MATCH_CLOZE_TAG       = regexp.MustCompile(`cloze:(\w+)`)
	REGEX_MATCH_CLOZE_FILTER    = regexp.MustCompile(`(?si)\{\{(?P<tag>c)(?P<ord>\d)+::(?P<content>.*?)(::(?P<hint>.*?))?\}\}`)
)

var RENDER_LIST = "list"

var RESET_COLOR = ansi.ColorCode("reset")

var HINT_COLOR = ansi.ColorCode("25")

var SPECIAL_FIELDS = []string{"Tags", "Type", "Deck", "Subdeck", "Card", "FrontSide"}

type FieldReplacmentOptions struct {
	card        models.Card
	isFrontSide bool
	isAnswer    bool
	readStruct  dynamicstruct.Reader
	cardTmpl    models.CardTemplate
	tmplFmt     string
	fieldMap    map[string]string
	renderType  string
}

// FieldReplacement
func FieldReplacement(opts FieldReplacmentOptions) (string, error) {
	var output strings.Builder
	var spVal string
	var err error

	filters := strings.Split(opts.tmplFmt, ":")
	// Check for simple field replacement or special field replacement
	if len(filters) == 1 {
		field := filters[0]
		raw, isField := opts.fieldMap[field]
		if isField {
			output.WriteString(fmt.Sprintf("{{.%s}}", raw))
		} else if IsSpecialFields(field) {
			// special fields will always reduce to a constant string
			// so include handlebars if not a special field
			spVal, err = SpecialFieldFilter(opts, field)
			if err != nil {
				return "", err
			}
			output.WriteString(fmt.Sprintf("%s ", spVal))
		} else {
			return "", fmt.Errorf("could not determine field type for %s", field)
		}
		return output.String(), nil
	}

	// Check for cloze deletion
	if REGEX_MATCH_CLOZE_TAG.MatchString(opts.tmplFmt) {
		clozeParts := REGEX_MATCH_CLOZE_TAG.FindStringSubmatch(opts.tmplFmt)
		if len(clozeParts) < 1 {
			return "", fmt.Errorf("could not find field from cloze tag %s", clozeParts)
		}

		raw, isField := opts.fieldMap[clozeParts[1]]
		if !isField {
			return "", fmt.Errorf("could not determine field for %s", clozeParts[1])
		}

		fieldValue := opts.readStruct.GetField(raw).String()
		clozeFilter, err := ReplaceWithClozeFilter(fieldValue, opts.card.Ord, opts.isAnswer)
		if err != nil {
			return "", err
		}
		output.WriteString(clozeFilter)
		return output.String(), nil
	}

	// Check for filter replacement
	// Filters will be deliminated by a `:` where the field is placed.
	// For instance, if filterA is applied to field_0 then the expected format will be
	// {{filterA:field_0}}
	output.WriteString("{{")
	field := filters[len(filters)-1]
	raw, isField := opts.fieldMap[field]
	if !isField {
		return "", fmt.Errorf("could not determine field")
	}
	for _, filter := range filters[0 : len(filters)-1] {
		filterWithArgs := strings.Split(filter, " ")
		output.WriteString(fmt.Sprintf("%s .%s", filterWithArgs[0], raw))
		if len(filterWithArgs) > 1 {
			output.WriteString(fmt.Sprintf(" %s", WrapArgsInQuotes(filterWithArgs[1:])))
		}

		if len(filters) > 2 {
			output.WriteString(" | ")
		}
	}
	output.WriteString("}}")

	return output.String(), nil
}

func IsSpecialFields(field string) bool {
	for _, sf := range SPECIAL_FIELDS {
		if sf == field {
			return true
		}
	}
	return false
}

func WrapArgsInQuotes(argList []string) string {
	// Use a counter to track argument while iterating
	ctr := len(argList) - 1
	return REGEX_MATCH_TMPL_ARG_GROUPS.ReplaceAllStringFunc(strings.Join(argList, " "), func(arg string) string {
		fmtStr := fmt.Sprintf(`"%s"`, strings.TrimSpace(arg))
		// Add a trailing space if not on the last argument
		if ctr > 0 {
			fmtStr = fmtStr + " "
		}
		ctr--
		return fmtStr

	})
}

func ReplaceWithClozeFilter(value string, cardOrd int, isAnswer bool) (string, error) {
	var err error = nil
	result := REGEX_MATCH_CLOZE_FILTER.ReplaceAllStringFunc(value, func(repl string) string {
		clozeValParts := REGEX_MATCH_CLOZE_FILTER.FindStringSubmatch(repl)
		// c, digit , text
		if len(clozeValParts) < 6 {
			return ""
		}
		// cloze "cloze ord" "card ord" "content" "isAnswer"
		return fmt.Sprintf("{{ cloze \"%s\" \"%d\" \"%s\" \"%d\" }}", clozeValParts[2], cardOrd+1, clozeValParts[3], utils.BoolToInt(isAnswer))
	})

	if result == "" {
		err = fmt.Errorf("could not create cloze filter from %s", value)
	}

	return result, err

}

func SpecialFieldFilter(opts FieldReplacmentOptions, field string) (string, error) {
	switch field {
	case "Tags":
		return opts.card.Note.StringTags, nil
	case "Type":
		return opts.card.Note.Model.Name, nil
	case "Subdeck":
		d := strings.Split(opts.card.Deck.Name, "::")
		return d[len(d)-1], nil
	case "Card":
		return opts.cardTmpl.Name, nil
	case "FrontSide":

		if !opts.isFrontSide {
			// generate the go template using front/question template
			return ParseCardTemplate(TemplateParseOptions{
				IsAnswer:         false,
				CardTemplateName: opts.cardTmpl.Name,
				CardTemplate:     opts.cardTmpl,
				Card:             opts.card,
				ReadStruct:       opts.readStruct,
				FieldMap:         opts.fieldMap,
				isFrontSide:      true,
			})
		}
		return "", fmt.Errorf("{{FrontSide}} only valid in back/answer template")
	}

	return "", fmt.Errorf("could not determine special field for field: %s", field)
}

func FieldReplacementMap(config *config.Config, renderType string) template.FuncMap {
	return template.FuncMap{
		"hint": func(field string) string {
			if renderType == RENDER_LIST {
				color := HINT_COLOR

				if config.Color.Hint != "" {
					color = ansi.ColorCode(config.Color.Hint)
				}
				field = strings.ReplaceAll(field, "[", "")
				field = strings.ReplaceAll(field, "]", "")
				return fmt.Sprintf("Show Extra: %s %s %s", color, field, RESET_COLOR)
			}
			return ""
		},
		"cloze": func(args ...string) (string, error) {
			if renderType == RENDER_LIST {
				color := ansi.ColorCode("blue")
				if len(args) < 3 {
					return "", fmt.Errorf("cloze requires at least cloze ord, card ord and content")
				}

				clozeOrd := args[0]
				cardOrd := args[1]
				content := args[2]
				isAnswer := args[3]
				if clozeOrd == cardOrd && isAnswer == "0" {
					content = "[...]"
				}
				if config.Color.Hint != "" {
					color = ansi.ColorCode(config.Color.Hint) // blue underline
				}
				return fmt.Sprintf("%s %s %s", color, content, RESET_COLOR), nil
			}
			return "", fmt.Errorf("could not determine how to render filter for type %v", renderType)
		},
		"furigana": func(field string) (string, error) {
			if renderType == RENDER_LIST {
				re := regexp.MustCompile(`[(.*)]`)
				kanjiField := re.ReplaceAllString(field, "")
				result, err := kakasi.Transform(kanjiField, kakasi.WithFurigana())
				result, err = kakasi.Transform(kanjiField, kakasi.WithWakatigaki())
				result, err = kakasi.Transform(kanjiField, kakasi.WithGraphic(kakasi.ASCII))
				if err != nil {
					return "", fmt.Errorf("furigana could not transform kanji into furigana: %+v", err.Error())
				}
				return result, nil
			}
			return "", fmt.Errorf("could not determine how to render filter for type %v", renderType)
		},
	}
}
