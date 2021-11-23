package create

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/aerex/anki-cli/api"
	_ "github.com/aerex/anki-cli/api/types"
	"github.com/aerex/anki-cli/api/types/rest"
	"github.com/aerex/anki-cli/internal/config"
	"github.com/aerex/anki-cli/pkg/anki"
	"github.com/aerex/anki-cli/pkg/io"
	"github.com/aerex/anki-cli/pkg/template"
	helpers "github.com/aerex/anki-cli/tests"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func executeCreateCommand(t *testing.T, cfg *config.Config, buffers *helpers.TestCmdBuffers, args []string, mockHttp func()) error {

  anki := &anki.Anki{
    Api: api.NewApi(cfg),
    Config: cfg,
    IO: io.NewTestIO(buffers.InBuf, buffers.OutBuf, buffers.ErrBuf),
    Templates: template.NewTemplate(cfg),
  }

  client := anki.Api.GetClient()
  httpmock.ActivateNonDefault(client)

  mockHttp()

  // Run deck cmd
  cmd := NewCreateCmd(anki, nil)
  if len(args) != 0 {
    cmd.SetArgs(args)
  }

  cmd.SetIn(buffers.InBuf)
  cmd.SetOut(buffers.OutBuf)
  cmd.SetErr(buffers.ErrBuf)

  err := cmd.Execute()

  return err
}

func TestCreateCreateUsingRest(t *testing.T) {
  // Setup
  name := "Default"
  var cfg config.Config
  if err := config.Load("../fixtures/", &cfg); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  testBufs := helpers.TestCmdBuffers{
    InBuf: &bytes.Buffer{},
    OutBuf: &bytes.Buffer{},
    ErrBuf : &bytes.Buffer{},
  }

  expectedOut := fmt.Sprintf("Created deck %s", name)

  args := []string{name}
  // Run command
  expectedUrl := fmt.Sprintf("%s%s", cfg.Endpoint, rest.DECKS_URI)
  err := executeCreateCommand(t, &cfg, &testBufs, args, func() {
    httpmock.RegisterResponder("POST", expectedUrl,
      httpmock.NewJsonResponderOrPanic(200, httpmock.File("../fixtures/decks/deck.json"),
      ))})
  defer httpmock.DeactivateAndReset()
  if err != nil {
    t.Errorf("Could not run the deck create command: %v", err)
  }
  httpmock.GetTotalCallCount()
  // Assertions
  assert.Equal(t, "", testBufs.ErrBuf.String())
  assert.Equal(t, expectedOut, testBufs.OutBuf.String())
  assert.Equal(t, 1, httpmock.GetCallCountInfo()[fmt.Sprintf("POST %s", expectedUrl)])
}
