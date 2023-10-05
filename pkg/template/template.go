package template

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/pkg/io"
	"github.com/aerex/go-anki/pkg/models"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/yaml.v2"
)

// TODO: Rename template names that make sense
// Don't do deck-list use list-deck
const (
	LIST_DECK               = "list-deck"
	DECK_SINGLE_OPTION_LIST = "deck-option-list"
	MULTIPLE_OPTIONS_LIST   = "deck-options-list"
	CARD_LIST               = "card-list"
	CREATE_CARD             = "create-card"
	LIST_CARD_TYPES         = "list-card-types"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Template
type Template interface {
	Load(string) error
	Execute(interface{}, *io.IO) error
	GetTemplateFile(pathname string) (string, error)
	LoadedTemplate() *template.Template
}

type AnkiTemplate struct {
	Config         *config.Config
	loadedTemplate *template.Template
	data           string
}

func NewTemplate(config *config.Config) Template {
	return &AnkiTemplate{
		Config: config,
	}
}
func LoadString(tmpl string, config *config.Config, renderType string) (*template.Template, error) {
	t, err := template.New("ankicli").Funcs(CustomFuncMaps()).
		Funcs(FieldReplacementMap(config, renderType)).
		Parse(tmpl)
	if err != nil {
		return nil, nil
	}
	return t, nil
}

// Return the full path of the sample template directory
func GetSampleTemplateFilePath(fileName string) string {
	_, moduleFileName, _, _ := runtime.Caller(0)
	moduleDir := filepath.Join(filepath.Dir(moduleFileName))
	return filepath.Join(moduleDir, "../../configs/templates", fileName)
}

// Read template file and return the contents
// Attempt to read from absolute path otherwise retrieve from
// User's templates config directory (ie $HOME/.config/templates)
func (t *AnkiTemplate) GetTemplateFile(pathname string) (string, error) {
	// TODO: Might want to either pass in filename as arg
	paths := strings.Split(pathname, "/")
	fileName := paths[len(paths)-1]

	templateConfigFilePath := filepath.Join(t.Config.Dir, "templates", fileName)
	var filePath string
	if _, err := os.Stat(pathname); os.IsNotExist(err) {
		// Check to see if the file exists in the template config dir
		if _, err = os.Stat(templateConfigFilePath); os.IsNotExist(err) {
			return "", err
		}
		filePath = templateConfigFilePath
	} else {
		filePath = pathname
	}

	// Read file and return data
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Load template for processing data.
// If template file cannot be found will try to read default
func (t *AnkiTemplate) Load(pathname string) error {
	if t.Config.Dir != "" {
		data, err := t.GetTemplateFile(pathname)
		if os.IsNotExist(err) {
			useDefault := false
			// Confirm with user that default template will be used.
			// Ask if cli should copy over sample template file
			survey.AskOne(
				&survey.Confirm{
					Message: "Could not find template file. Would you like to use the default file?",
					Help:    "You can run anki templates generate to copy sample template files to your anki-cli config directory",
				}, &useDefault, nil)

			if useDefault {
				// If we get to this point we can assume pathname is a fileName
				// since  we could not find template override provided by the user
				data, err = t.GetTemplateFile(GetSampleTemplateFilePath(pathname))
				if err != nil {
					return err
				}
			} else {
				// NOTE: This is fine but there should a better output message we can return
				// or use no error
				return fmt.Errorf("there was a problem loading the template file %v", pathname)
			}
		}

		// Load template engine with useful template functions
		t.loadedTemplate = template.New("ankicli").Funcs(CustomFuncMaps())
		t.data = data
		return nil
	}
	return fmt.Errorf("there was a problem loading the template file %v", pathname)
}

func (t *AnkiTemplate) Execute(data interface{}, io *io.IO) error {
	if t.loadedTemplate != nil {
		tmpl, err := t.loadedTemplate.Funcs(TableFuncMap(io)).Funcs(CustomFuncMaps()).Parse(t.data)
		if err != nil {
			return err
		}
		err = tmpl.Execute(io.Output, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *AnkiTemplate) LoadedTemplate() *template.Template {
	return t.loadedTemplate
}

// dateInZone
// see sprig
func dateInZone(fmt string, date interface{}, zone string) (string, error) {
	var t time.Time
	switch date := date.(type) {
	default:
		t = time.Now()
	case time.Time:
		t = date
	case *time.Time:
		t = *date
	case models.UnixTime:
		t = time.Unix(int64(date), 0)
	case int64:
		t = time.Unix(date, 0)
	case int:
		t = time.Unix(int64(date), 0)
	case int32:
		t = time.Unix(int64(date), 0)
	}

	loc, err := time.LoadLocation(zone)
	if err != nil {
		loc, err = time.LoadLocation("UTC")
		if err != nil {
			return "", err
		}
	}

	return t.In(loc).Format(fmt), nil
}

func CustomFuncMaps() template.FuncMap {
	return template.FuncMap{
		"toJson": func(content interface{}) (string, error) {
			jsonData, err := json.Marshal(content)
			if err != nil {
				return "", err
			}
			return string(jsonData), nil
		},
		"toYaml": func(content interface{}) (string, error) {
			yamlData, err := yaml.Marshal(content)
			if err != nil {
				return "", err
			}
			return string(yamlData), nil
		},
		"loop": func(count int) []int {
			var i int
			var items []int
			for i = 0; i < (count); i++ {
				items = append(items, i)
			}
			return items
		},
		"date": func(fmt string, content interface{}) (string, error) {
			return dateInZone(fmt, content, "Local")
		},
	}
}
