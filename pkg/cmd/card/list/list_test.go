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
		Api:       api.NewApi(cfg),
		Config:    cfg,
		IO:        io.NewTestIO(buffers.InBuf, buffers.OutBuf, buffers.ErrBuf),
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
	headers := []string{"Question", "Answer", "Card Type", "Deck", "Due Date"}
	data := [][]string{
		{"Q1", "A1", "Basic with sentences", "Default", "2021-11-22"},
		{"Q2", "A2", "Basic with sentences", "Basic", "2021-11-22"},
	}
	args := []string{"-t", "./fixtures/templates/card-list"}
	expectedTableOut := helpers.GenerateTableOutputWithHeaders(headers, data)

	// Run command
	err = executeListCommand(t, &cfg, &testBufs, args, func() {
		httpmock.RegisterResponder("GET", cfg.Endpoint+rest.CARDS_URI,
			httpmock.NewJsonResponderOrPanic(200, httpmock.File("./fixtures/cards/cards.json")))
	})
	defer httpmock.DeactivateAndReset()

	if err != nil {
		t.Errorf("Could not run the cards list command: %v", err)
	}

	// Assertions
	assert.Equal(t, "", testBufs.ErrBuf.String())
	assert.Contains(t, expectedTableOut, testBufs.OutBuf.String())
}

func TestDefaultListQueryUsingRestCmd(t *testing.T) {
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
	headers := []string{"Question", "Answer", "Card Type", "Deck", "Due Date"}
	data := [][]string{
		{"Q2", "A2", "Basic with sentences", "Basic", "2021-11-22"},
	}
	expectedQuery := "deck:Basic"
	args := []string{"-t", "./fixtures/templates/card-list", "-q", expectedQuery}
	expectedTableOut := helpers.GenerateTableOutputWithHeaders(headers, data)

	// Run command
	err = executeListCommand(t, &cfg, &testBufs, args, func() {
		httpmock.RegisterResponderWithQuery("GET", cfg.Endpoint+rest.CARDS_URI,
			url.Values{"query": []string{expectedQuery}},
			httpmock.NewJsonResponderOrPanic(200, httpmock.File("./fixtures/cards/card.json")))
	})
	defer httpmock.DeactivateAndReset()

	if err != nil {
		t.Errorf("Could not run the cards list command: %v", err)
	}

	// Assertions
	assert.Equal(t, "", testBufs.ErrBuf.String())
	assert.Contains(t, expectedTableOut, testBufs.OutBuf.String())
}
