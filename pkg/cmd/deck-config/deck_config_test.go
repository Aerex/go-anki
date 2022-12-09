package deck_config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aerex/go-anki/api"
	"github.com/aerex/go-anki/api/rest"
	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/io"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/pkg/template"
	helpers "github.com/aerex/go-anki/tests"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func generateSingleDeckConfigYamlString(path string) (string, error) {
	model := models.DeckConfig{}
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
func executeDeckConfigsCommand(t *testing.T, cfg *config.Config, buffers *helpers.TestCmdBuffers, args []string, mockHttp func()) error {

	anki := &anki.Anki{
		Api:       api.NewApi(cfg),
		Config:    cfg,
		IO:        io.NewTestIO(buffers.InBuf, buffers.OutBuf, buffers.ErrBuf, nil),
		Templates: template.NewTemplate(cfg),
	}

	client := anki.Api.GetClient()
	httpmock.ActivateNonDefault(client)

	mockHttp()

	// Run deck cmd
	cmd := NewDeckConfigsCmd(anki, nil)
	if len(args) != 0 {
		cmd.SetArgs(args)
	}

	cmd.SetIn(buffers.InBuf)
	cmd.SetOut(buffers.OutBuf)
	cmd.SetErr(buffers.ErrBuf)

	err := cmd.Execute()

	return err
}
func TestGetSingleDeckConfigUsingRest(t *testing.T) {
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
	expectedDeckConfigsUrl := fmt.Sprintf("GET %s%s/%s", cfg.Endpoint, rest.DECK_CONFIGS_URI, optionName)

	// Run command
	expectedUrl := fmt.Sprintf("%s%s/%s", cfg.Endpoint, rest.DECK_CONFIGS_URI, optionName)
	err = executeDeckConfigsCommand(t, &cfg, &testBufs, []string{optionName}, func() {
		httpmock.RegisterResponder("GET", expectedUrl,
			httpmock.NewJsonResponderOrPanic(200, httpmock.File("./fixtures/single_option.json")))
	})
	if err != nil {
		t.Errorf("Could not run the deck options command: %v", err)
	}
	defer httpmock.DeactivateAndReset()
	expectedOut, err := generateSingleDeckConfigYamlString("./fixtures/single_option.json")
	if err != nil {
		t.Errorf("Could not run the deck options command: %v", err)
	}

	assert.Equal(t, "", testBufs.ErrBuf.String())
	assert.Contains(t, testBufs.OutBuf.String(), expectedOut)
	assert.Equal(t, 1, httpmock.GetCallCountInfo()[expectedDeckConfigsUrl])
}

// FIXME: Redo this once it is determine how edits for deck configs will work
// See deck_configs.go#NewDeckConfigsCmd for more information
//func TestEditSingleDeckConfigWithEditorUsingRest(t *testing.T) {
//	// Setup
//	optionName := "Default"
//	cfg, err := config.LoadSampleConfig()
//	if err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//	testBufs := helpers.TestCmdBuffers{
//		InBuf:  &bytes.Buffer{},
//		OutBuf: &bytes.Buffer{},
//		ErrBuf: &bytes.Buffer{},
//	}
//	httpmock.GetTotalCallCount()
//
//	// Run command
//	expectedGetDeckConfigsUrl := fmt.Sprintf("GET %s%s/%s", cfg.Endpoint, rest.DECK_CONFIGS_URI, optionName)
//	expectedUpdateDeckConfigsUrl := fmt.Sprintf("PATCH %s%s/%s", cfg.Endpoint, rest.DECK_CONFIGS_URI, optionName)
//	err = executeDeckConfigsCommand(t, &cfg, &testBufs, []string{optionName}, func() {
//		httpmock.RegisterResponder("GET", expectedGetDeckConfigsUrl,
//			httpmock.NewJsonResponderOrPanic(200, httpmock.File("./fixtures/single_option.json")))
//		httpmock.RegisterResponder("PATCH", expectedUpdateDeckConfigsUrl,
//			httpmock.NewJsonResponderOrPanic(200, httpmock.File("./fixtures/single_option.json")))
//	})
//	if err != nil {
//		t.Errorf("Could not run the deck options command: %v", err)
//	}
//	defer httpmock.DeactivateAndReset()
//	expectedOut, err := generateSingleDeckConfigYamlString("./fixtures/single_option.json")
//	if err != nil {
//		t.Errorf("Could not run the deck options command: %v", err)
//	}
//
//	assert.Equal(t, "", testBufs.ErrBuf.String())
//	assert.Contains(t, testBufs.OutBuf.String(), expectedOut)
//	assert.Equal(t, 1, httpmock.GetCallCountInfo()[expectedGetDeckConfigsUrl])
//	assert.Equal(t, 1, httpmock.GetCallCountInfo()[expectedUpdateDeckConfigsUrl])
//
//}
