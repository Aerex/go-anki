package anki

import (
	"github.com/aerex/anki-cli/api"
	"github.com/aerex/anki-cli/internal/config"
	"github.com/aerex/anki-cli/pkg/io"
	"github.com/aerex/anki-cli/pkg/template"
	"github.com/op/go-logging"
)

type Anki struct {
  Api api.Api
  IO *io.IO
  Config *config.Config
  Templates template.Template
  Log *logging.Logger
}
