package logger

import (
	"os"
	"strings"

	"github.com/aerex/anki-cli/pkg/anki"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

func ConfigureLogger(anki *anki.Anki, module string) error {
  var logBackend logging.Backend
  if anki.Config.Logger.File != "" {
    f, err := os.OpenFile(anki.Config.Logger.File, os.O_APPEND | os.O_CREATE | os.O_RDWR, 0666)
    if err != nil {
      return err
    }
    // Save ref to file buffer
    anki.IO.Log = f

    logBackend = logging.NewLogBackend(f, "", 0)
  } else {
    logBackend = logging.NewLogBackend(anki.IO.Error, "", 0)
  }

  // Get log format from config
  format := viper.GetString("logger.format")
  if format == "" {
    format = "%{color}%{time:2006-01-02T15:04:05}%{level:-4s}%{color:reset} %{message}"
  }
  logFormat := logging.NewBackendFormatter(logBackend, logging.MustStringFormatter(format))

  level := strings.ToUpper(viper.GetString("logger.level"))
  if level == "" {
    level = logging.DEBUG.String()
  }
  logLevel := logging.AddModuleLevel(logBackend)
  l, err := logging.LogLevel(level)
  if err != nil {
    return err
  }
  logLevel.SetLevel(l, module)

  logging.SetBackend(logLevel, logFormat)
  // TOD: Add multi log backends to output to both file and  to 
  return  nil
}
