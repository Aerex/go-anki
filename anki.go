package main

import (
	"fmt"
	"os"

	"github.com/aerex/go-anki/api"
	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/internal/logger"
	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/editor"
	"github.com/aerex/go-anki/pkg/io"
	"github.com/aerex/go-anki/pkg/root"
	"github.com/aerex/go-anki/pkg/template"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"

	// Preload all api types
	_ "github.com/aerex/go-anki/api/types"
)

func main() {
	var cfg config.Config
	i := io.NewSystemIO()

	anki := &anki.Anki{
		IO: i,
	}

	// PERF: Maybe move this if block into a method
	if err := config.Load("", &cfg, anki.IO); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// If we could not find the config ask the user if we can copy over the sample
			createConfigFile := false
			survey.AskOne(
				&survey.Confirm{
					Message: "A configuration file could not be found. Would you like a sample configuration created for you?",
				}, &createConfigFile, nil)

			// If yes, copy over sample to user configuration directory
			if createConfigFile {
				configFilePath, err := config.GenerateSampleConfig(&cfg, io.NewSystemIO())
				if err != nil {
					fmt.Fprintln(os.Stderr, "Could not generate sample configuration file ", err)
					os.Exit(1)
				}
				fmt.Fprintln(os.Stdout, "Generated configuration file in ", configFilePath)
			} else {
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	anki.Config = &cfg
	anki.Templates = template.NewTemplate(&cfg)
	anki.Editor = editor.NewModelEditor(anki.Templates, io.NewSystemIO())

	log, err := logger.ConfigureLogger(anki)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	anki.Api = api.NewApi(&cfg, log)

	// Run anki-cli
	var root = root.NewRootCmd(anki)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
