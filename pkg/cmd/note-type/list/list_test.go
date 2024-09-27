package list

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/aerex/go-anki/api"
	"github.com/aerex/go-anki/api/rest"
	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/io"
	"github.com/aerex/go-anki/pkg/template"
	helpers "github.com/aerex/go-anki/tests"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func generateCardTypeYamlString(path string) (string, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(file), "\\n", "\n"), nil
}

// TODO: generalize this method and the other into one
func executeListCommand(t *testing.T, cfg *config.Config, buffers *helpers.TestCmdBuffers, args []string, mockHttp func()) error {

	anki := &anki.Anki{
		API:       api.NewApi(cfg),
		Config:    cfg,
		IO:        io.NewTestIO(buffers.InBuf, buffers.OutBuf, buffers.ErrBuf, nil),
		Templates: template.NewTemplate(cfg),
	}

	client := anki.API.GetClient()
	httpmock.ActivateNonDefault(client)

	mockHttp()

	// Run card-types cmd
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
func TestGetSingleCardTypeUsingRest(t *testing.T) {
	// Setup
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
	cardType := "Basic"

	// Run command
	args := []string{"-t", "./fixtures/templates/list-card-types", "-n", cardType}
	expectedUrl := fmt.Sprintf("GET %s%s/models?name=%s", cfg.Endpoint, rest.COLLECTION_URI, cardType)
	err = executeListCommand(t, &cfg, &testBufs, args, func() {
		httpmock.RegisterResponderWithQuery("GET", cfg.Endpoint+rest.COLLECTION_URI+"/models",
			url.Values{"name": []string{cardType}},
			httpmock.NewJsonResponderOrPanic(200, httpmock.File("./fixtures/card-type.json")))
	})
	if err != nil {
		t.Errorf("Could not run the card types command: %v", err)
	}
	defer httpmock.DeactivateAndReset()
	expectedOut, err := generateCardTypeYamlString("./fixtures/card-type.yaml")
	if err != nil {
		t.Errorf("Could not run the card types command: %v", err)
	}

	assert.Equal(t, "", testBufs.ErrBuf.String())
	assert.Contains(t, testBufs.OutBuf.String(), expectedOut)
	assert.Equal(t, 1, httpmock.GetCallCountInfo()[expectedUrl])
}
