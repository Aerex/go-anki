package logger

import (
	"os"
	"strings"
	"time"

	"github.com/aerex/go-anki/pkg/anki"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var logger zerolog.Logger

func ConfigureLogger(anki *anki.Anki) (*zerolog.Logger, error) {
	if anki.Config.Logger.File != "" {
		path, err := homedir.Expand(anki.Config.Logger.File)
		if err != nil {
			return nil, err
		}

		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}

    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        f,
			NoColor:    true,
			TimeFormat: time.RFC3339,
		})
	} else {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out: anki.IO.Error,
		})
	}

	level := strings.ToUpper(viper.GetString("logger.level"))
	switch level {
	case "TRACE":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERR":
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "FATAL":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

  format := viper.GetString("logger.format")
  if format != "" {
  }
	// Save ref to file buffer
	return &logger, nil
}
