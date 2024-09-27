package deck_config

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/MakeNowJust/heredoc"
	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/pkg/template"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type DeckConfigOptions struct {
	Edit     bool
	Template string
}

// TODO: Figure out how to edit deck-configs using a human-usble ref id
// Current id is an timestamp
func NewDeckConfigsCmd(anki *anki.Anki, overrideF func(*anki.Anki) error) *cobra.Command {

	opts := &DeckConfigOptions{}
	cmd := &cobra.Command{
		Use:   "options [deck_name]",
		Short: "Show or edit deck configs/options",
		Example: heredoc.Doc(`
      $ anki deck-configs Default
      $ anki deck options Default --edit
      $ anki deck options Default group Default
      $ anki deck options Default new.steps 110
    `),
		RunE: func(cmd *cobra.Command, args []string) error {
			if overrideF != nil {
				return overrideF(anki)
			}
			return deckConfigsCmd(anki, args, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Edit, "edit", "e", false, "Edit one or multiple deck options in your editor")
	cmd.Flags().StringVarP(&opts.Template, "template", "t", "", "Override template for output")

	// TODO:
	// Add --edit option to open an editor to configure options as yaml (add a limit of 5 if not editing just one)
	// Add the ability to update nested options like new.steps
	// Add the ability to see all options using default template (deck-options) or given template

	return cmd
}

func deckConfigsCmd(anki *anki.Anki, args []string, opts *DeckConfigOptions) error {

	// TODO: Need to generalize scenarios such as one deck option
	// scenarios are currently:
	// - get one deck config,
	// - get all deck configs,
	// - use editor on deck config,
	// - get an option key-value pair
	// - get an nested option  key-value pair
	// Handling one deck option
	if len(args) == 1 {
		tmpl := template.DECK_SINGLE_OPTION_LIST
		if opts.Template != "" {
			tmpl = opts.Template
		}
		if err := anki.Templates.Load(tmpl); err != nil {
			return err
		}
		options, err := anki.API.GetDeckConfig(args[0])
		if err != nil {
			return err
		}
		if opts.Edit {
			file, err := editOptions(anki, options)
			if err != nil {
				return err
			}
			updatedOption := models.DeckConfig{}
			yaml.Unmarshal(file, &updatedOption)

			// TODO: Plan to print updated deck config into logger
			_, err = anki.API.UpdateDeckConfig(updatedOption, "")
			if err != nil {
				return err
			}

		} else {
			if err := anki.Templates.Execute(options, anki.IO); err != nil {
				return err
			}
		}
	}
	return nil
}

func getEditor() string {
	editor := viper.GetString("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			return "notepad.exe"
		}
		return "vim"
	}
	return editor
}

// TODO: Move to io package later
func makeTmpEditFile() (string, *bytes.Buffer, error) {
	file, err := os.CreateTemp("", "ankicli-edit-*.yaml")
	if err != nil {
		return "", &bytes.Buffer{}, err
	}
	return file.Name(), &bytes.Buffer{}, err
}

// TODO: Move to io package later
func editOptions(anki *anki.Anki, data interface{}) ([]byte, error) {
	editorCmd := getEditor()
	originalOut := anki.IO.Output
	editFilePath, editBuf, err := makeTmpEditFile()
	if err != nil {
		return []byte{}, err
	}
	// Make duplicate buffer
	var copyEditBuf bytes.Buffer
	editWriter := io.MultiWriter(editBuf, &copyEditBuf)
	anki.IO.Output = editWriter
	// Cleanup after editing
	defer func() {
		anki.IO.Output = originalOut
		os.RemoveAll(editFilePath)
	}()

	if err := anki.Templates.Execute(data, anki.IO); err != nil {
		return []byte{}, err
	}

	cmd := exec.Command(editorCmd, editFilePath)
	cmd.Stdout, cmd.Stdin, cmd.Stderr = originalOut, anki.IO.Input, anki.IO.Error
	if err := cmd.Run(); err != nil {
		return []byte{}, err
	}

	file, err := os.ReadFile(editFilePath)
	if err != nil {
		return []byte{}, err
	}
	// TODO: Add validation on edit and allow option to loop until edit is valid (use yaml.Unmarshal to validate)
	// TODO: Add check to see if there are any changes. Only push when there are changes

	return file, nil
}
