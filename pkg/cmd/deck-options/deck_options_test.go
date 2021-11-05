package options

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aerex/anki-cli/api"
	"github.com/aerex/anki-cli/api/types/rest"
	"github.com/aerex/anki-cli/internal/config"
	"github.com/aerex/anki-cli/pkg/anki"
	"github.com/aerex/anki-cli/pkg/io"
	"github.com/aerex/anki-cli/pkg/models"
	"github.com/aerex/anki-cli/pkg/template"
	helpers "github.com/aerex/anki-cli/tests"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)


func generateYamlOutputFromJSONFile(path string) (string, error) {
  model := []models.DeckOptions{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(file, &model)
	if err != nil {
		return "", err
	}

	yamlData, err := yaml.Marshal(model)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

// TODO: generalize this method and the other into one
func executeOptionsCommand(t *testing.T, cfg *config.Config, buffers *helpers.TestCmdBuffers, args []string, mockHttp func()) error {

	anki := &anki.Anki{
		Api:       api.NewApi(cfg),
		Config:    cfg,
		IO:        io.NewTestIO(buffers.InBuf, buffers.OutBuf, buffers.ErrBuf),
		Templates: template.NewTemplate(cfg),
	}

	client := anki.Api.GetClient()
	httpmock.ActivateNonDefault(client)

	mockHttp()

	// Run deck cmd
	cmd := NewOptionsCmd(anki, nil)
	if len(args) != 0 {
		cmd.SetArgs(args)
	}

	cmd.SetIn(buffers.InBuf)
	cmd.SetOut(buffers.OutBuf)
	cmd.SetErr(buffers.ErrBuf)

	err := cmd.Execute()

	return err
}
func TestGetSingleDeckOptionsUsingRest(t *testing.T) {
	// Setup
	optionName := "Default"
	cfg, err := config.LoadSampleConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	testBufs := helpers.TestCmdBuffers{
		InBuf:  &bytes.Buffer{},
		OutBuf: &bytes.Buffer{},
		ErrBuf: &bytes.Buffer{},
	}
  httpmock.GetTotalCallCount()
  expectedDeckOptionsUrl := fmt.Sprintf("GET %s%s/%s", cfg.Endpoint, rest.DECK_OPTIONS_URI, optionName)

	// Run command
	expectedUrl := fmt.Sprintf("%s%s/%s", cfg.Endpoint, rest.DECK_OPTIONS_URI, optionName)
	err = executeOptionsCommand(t, &cfg, &testBufs, []string{optionName}, func() {
		httpmock.RegisterResponder("GET", expectedUrl,
			httpmock.NewJsonResponderOrPanic(200, httpmock.File("./fixtures/all_options.json")))
	})
	if err != nil {
		t.Errorf("Could not run the deck options command: %v", err)
	}
	defer httpmock.DeactivateAndReset()
  expectedOut, err := generateYamlOutputFromJSONFile("./fixtures/all_options.json")
  if err != nil {
		t.Errorf("Could not run the deck options command: %v", err)
  }

  assert.Equal(t, "", testBufs.ErrBuf.String())
  assert.Contains(t,  testBufs.OutBuf.String(),expectedOut)
  assert.Equal(t, 1, httpmock.GetCallCountInfo()[expectedDeckOptionsUrl])
}
