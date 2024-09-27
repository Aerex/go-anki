package prompt

import (
	"os"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/aerex/go-anki/internal/config"
)

type surveyPrompt struct {
	config config.Config
}

func NewSurveyPrompt(cfg config.Config) Prompt {
	return &surveyPrompt{
		config: cfg,
	}
}

func (s surveyPrompt) Choose(title string, options []string, defaultOption string) (answers string, err error) {
	err = survey.AskOne(&survey.Select{
		Message: title,
		Options: options,
		Default: defaultOption,
		VimMode: s.config.Prompt.Vim,
	}, &answers)
	defer Interrupt(err)
	return
}

func (s surveyPrompt) Confirm(title string) (confirm bool, err error) {
	err = survey.AskOne(&survey.Confirm{
		Message: title,
	}, &confirm)
	defer Interrupt(err)
	return
}

func (s surveyPrompt) Select(title string, options []string, defaultOpt string) (answers []string, err error) {
	err = survey.AskOne(&survey.MultiSelect{
		Message: title,
		Options: options,
		Default: defaultOpt,
		VimMode: s.config.Prompt.Vim,
	}, &answers)
	defer Interrupt(err)
	return
}

func Interrupt(err error) {
	if err == terminal.InterruptErr {
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}
}
