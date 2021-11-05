package template

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/aerex/anki-cli/internal/config"
	"github.com/aerex/anki-cli/pkg/io"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/yaml.v2"
)

const (
	DECK_LIST = "deck-list"
  DECK_SINGLE_OPTION_LIST = "deck-option-list"
  MULTIPLE_OPTIONS_LIST = "deck-options-list"
)

type Template interface {
	Load(string) error
	Execute(interface{}, *io.IO) error
	GetTemplateFile(pathname string) (string, error)
}

type AnkiTemplate struct {
	Config         *config.Config
	LoadedTemplate *template.Template
	data           string
}

func NewTemplate(config *config.Config) Template {
	return &AnkiTemplate{
		Config: config,
	}
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
	templateConfigFilePath := filepath.Join(t.Config.Dir, "templates", pathname)
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
			}
		}

		// Load template engine with useful template functions
		t.LoadedTemplate = template.New("ankicli").Funcs(sprig.GenericFuncMap())
		t.data = data
		return nil
	}
	return fmt.Errorf("there was a problem loading the template file %v", pathname)
}

func (t *AnkiTemplate) Execute(data interface{}, io *io.IO) error {
	if t.LoadedTemplate != nil {
		tmpl, err := t.LoadedTemplate.Funcs(TableFuncMap(io)).Funcs(CustomFuncMaps()).Parse(t.data)
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

func CustomFuncMaps() map[string]interface{} {
	return map[string]interface{}{
		"toYaml": func(content interface{}) (string, error) {
			yamlData, err := yaml.Marshal(content)
			if err != nil {
				return "", err
			}
			return string(yamlData), nil
		},
	}
}
