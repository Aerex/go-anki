package list

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/aerex/anki-cli/api"
  helpers "github.com/aerex/anki-cli/tests"
	_ "github.com/aerex/anki-cli/api/types"
	"github.com/aerex/anki-cli/api/types/rest"
	"github.com/aerex/anki-cli/internal/config"
	"github.com/aerex/anki-cli/pkg/anki"
	"github.com/aerex/anki-cli/pkg/io"
	"github.com/aerex/anki-cli/pkg/template"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestDefaultListUsingRestCmd(t *testing.T) {
  // Setup

  var cfg config.Config
  if err := config.Load("../fixtures/", &cfg); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  inBuf := &bytes.Buffer{} 
  outBuf := &bytes.Buffer{} 
  errBuf := &bytes.Buffer{} 

  anki := &anki.Anki{
    Api: api.NewApi(&cfg),
    Config: &cfg,
    IO: io.NewTestIO(inBuf, outBuf, errBuf),
    Templates: template.NewTemplate(&cfg),
  }

  client := anki.Api.GetClient()
  httpmock.ActivateNonDefault(client)
  defer httpmock.DeactivateAndReset()
  
  httpmock.RegisterResponder("GET", cfg.Endpoint +  rest.DECKS_URI,
    httpmock.NewJsonResponderOrPanic(200, httpmock.File("../fixtures/decks.json"),
  ))

  // Run deck cmd
  cmd := NewListCmd(anki, nil)

  cmd.SetIn(inBuf)
  cmd.SetOut(outBuf)
  cmd.SetErr(errBuf)


  err := cmd.Execute()
  if err != nil {
    t.Errorf("Could not run the decks list command: %v", err)
  }
  
  headers := []string{"Name", "Due", "Next"}
  data := [][]string{
   {"Default", "1", "5"},
   {"Kanji", "5", "2"},
  }

  expectdOut := helpers.GenerateTableOutputWithHeaders(headers, data)
  
  // Assertions
  assert.Equal(t, "", errBuf.String()) 
  assert.Equal(t, expectdOut, outBuf.String())
}
