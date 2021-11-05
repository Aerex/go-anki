package list

import (
	"bytes"
	"fmt"
	"net/url"
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

func executeListCommand(t *testing.T, cfg *config.Config, buffers *helpers.TestCmdBuffers, args []string, mockHttp func()) error {

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
  cmd := NewListCmd(anki, nil)
  if len(args) != 0 {
    cmd.SetArgs(args)
  }

  cmd.SetIn(buffers.InBuf)
  cmd.SetOut(buffers.OutBuf)
  cmd.SetErr(buffers.ErrBuf)

  err := cmd.Execute()

  return err
}

func TestDefaultListUsingRestCmd(t *testing.T) {
  // Setup
  var cfg config.Config
  if err := config.Load("../fixtures/", &cfg); err != nil {
    fmt.Println(err)
    os.Exit(1) }
  testBufs := helpers.TestCmdBuffers{
    InBuf: &bytes.Buffer{},
    OutBuf: &bytes.Buffer{},
    ErrBuf : &bytes.Buffer{},
  }
  headers := []string{"Name", "Due", "Next"}
  data := [][]string{
   {"Default", "1", "5"},
   {"Kanji", "5", "2"},
  }
  expectedTableOut := helpers.GenerateTableOutputWithHeaders(headers, data)
  expectedFooterOut := "Studied 10 cards in 100 seconds"

  // Run command
  err := executeListCommand(t, &cfg, &testBufs, []string{}, func() {
      httpmock.RegisterResponder("GET", cfg.Endpoint +  rest.DECKS_URI,
        httpmock.NewJsonResponderOrPanic(200, httpmock.File("../fixtures/decks/decks.json"),
      ))
      httpmock.RegisterResponderWithQuery("GET", cfg.Endpoint +  rest.COLLECTION_URI,
        url.Values{"include": []string{"meta"}},
        httpmock.NewJsonResponderOrPanic(200, httpmock.File("../fixtures/collection/meta.json"),
      ))
  })
  defer httpmock.DeactivateAndReset()

  if err != nil {
    t.Errorf("Could not run the decks list command: %v", err)
  }

  // Assertions
  assert.Equal(t, "", testBufs.ErrBuf.String())
  assert.Contains(t, testBufs.OutBuf.String(), expectedTableOut)
  assert.Contains(t, testBufs.OutBuf.String(), expectedFooterOut)
}

func TestExactMatchFilterDecksRestCmd(t *testing.T) {
  // Setup
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
  headers := []string{"Name", "Due", "Next"}
  data := [][]string{
    {"Default", "1", "5"},
  }
  expectedQuery := "deck:Default"
  args := []string{
    "--query=deck:Default",
  }
  expectedFooterOut := "Studied 10 cards in 100 seconds"
  expectedTableOut := helpers.GenerateTableOutputWithHeaders(headers, data)
  expectedGetDeckUrl := fmt.Sprintf("GET %s%s?query=%s", cfg.Endpoint, rest.DECKS_URI, url.QueryEscape(expectedQuery))
  expectedGetColMetaUrl := fmt.Sprintf("GET %s%s?include=%s&query=%s", cfg.Endpoint, rest.COLLECTION_URI, url.QueryEscape("meta"),
    url.QueryEscape(expectedQuery))
  httpmock.GetTotalCallCount()

  // Run command
  err := executeListCommand(t, &cfg, &testBufs, args, func() {
    httpmock.RegisterResponderWithQuery("GET", cfg.Endpoint + rest.DECKS_URI,
      url.Values{"query": []string{expectedQuery}},
      httpmock.NewJsonResponderOrPanic(200, httpmock.File("../fixtures/decks/single_deck.json"),
      ))
    httpmock.RegisterResponderWithQuery("GET", cfg.Endpoint +  rest.COLLECTION_URI,
      url.Values{"include": []string{"meta"}, "query": []string{expectedQuery}},
      httpmock.NewJsonResponderOrPanic(200, httpmock.File("../fixtures/collection/meta.json"),
      ))
  })
  defer httpmock.DeactivateAndReset()

  if err != nil {
    t.Errorf("Could not run the decks list command: %v", err)
  }

  // Assertions
  assert.Equal(t, "", testBufs.ErrBuf.String())
  assert.Contains(t, testBufs.OutBuf.String(), expectedTableOut)
  assert.Contains(t, testBufs.OutBuf.String(), expectedFooterOut)
  assert.Equal(t, 1, httpmock.GetCallCountInfo()[expectedGetDeckUrl])
  assert.Equal(t, 1, httpmock.GetCallCountInfo()[expectedGetColMetaUrl])
}
