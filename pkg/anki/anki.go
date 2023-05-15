package anki

import (
	"github.com/aerex/go-anki/api"
	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/pkg/editor"
	"github.com/aerex/go-anki/pkg/io"
	"github.com/aerex/go-anki/pkg/template"
	"github.com/op/go-logging"
)

type Anki struct {
	Api       api.Api
	IO        *io.IO
	Config    *config.Config
	Templates template.Template
	Log       *logging.Logger
	Editor    editor.Editor
}
