package main

import (
	"fmt"
	"os"

	"github.com/aerex/anki-cli/api"
	"github.com/aerex/anki-cli/internal/config"
	"github.com/aerex/anki-cli/internal/logger"
	"github.com/aerex/anki-cli/pkg/anki"
	"github.com/aerex/anki-cli/pkg/io"
	"github.com/aerex/anki-cli/pkg/root"
	"github.com/aerex/anki-cli/pkg/template"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"

	// Preload all api types
	_ "github.com/aerex/anki-cli/api/types"
)


func main() {
  var cfg config.Config

  if err := config.Load("", &cfg); err != nil {
    if _, ok := err.(viper.ConfigFileNotFoundError); ok {
      // If we could not find the config ask the user if we can copy over the sample 
      createConfigFile := false
      survey.AskOne(
        &survey.Confirm{
        Message: "A configuration file could not be found. Would you like a sample configuration created for you?",
      }, &createConfigFile, nil)

      // If yes, copy over sample to user configuration directory
      if createConfigFile {
        configFilePath, err := config.GenerateSampleConfig(&cfg)
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

  anki := &anki.Anki{
    Api: api.NewApi(&cfg),
    Config: &cfg,
    IO: io.NewSystemIO(),
    Templates: template.NewTemplate(&cfg),
  }
  if err := logger.ConfigureLogger(anki, "ankicli"); err != nil {
    fmt.Println(err)
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }
  if anki.IO.Log != nil {
    defer anki.IO.Log.Close()
  }
  var log = logging.MustGetLogger("ankicli")
  // Store ref to global logger
  anki.Log = log

  // Run anki-cli
  var root = root.NewRootCmd(anki)
  if err := root.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
