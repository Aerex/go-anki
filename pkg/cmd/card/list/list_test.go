package list

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/aerex/go-anki/api"
	"github.com/aerex/go-anki/api/rest"
	_ "github.com/aerex/go-anki/api/types"
	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/internal/utils"
	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/io"
	"github.com/aerex/go-anki/pkg/template"
	helpers "github.com/aerex/go-anki/tests"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func executeListCommand(t *testing.T, useHttp bool, cfg *config.Config, buffers *helpers.TestCmdBuffers, args []string, mockHttp func()) error {

	anki := &anki.Anki{
		Api:       api.NewApi(cfg),
		Config:    cfg,
		IO:        io.NewTestIO(buffers.InBuf, buffers.OutBuf, buffers.ErrBuf, exec.Command),
		Templates: template.NewTemplate(cfg),
	}

	if useHttp {
		client := anki.Api.GetClient()
		httpmock.ActivateNonDefault(client)
		mockHttp()
	}

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
	// TODO: Figure out why the due date is always printing today's date
	data := [][]string{
		{"A1a", "A1b", "Basic with sentences", "Default", "2021-11-22"},
		{"A2a", "A2b", "Basic with sentences", "Basic", "2021-11-22"},
	}
	args := []string{"-t", "./fixtures/templates/card-list"}
	expectedTableOut := helpers.GenerateTableOutputWithHeaders(headers, data)

	// Run command
	err = executeListCommand(t, true, &cfg, &testBufs, args, func() {
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

func TestDefaultListUsingSqlCmd(t *testing.T) {
	// Setup
	dbDir := filepath.Join(utils.CurrentModuleDir(), "/fixtures/db/collection.anki2")
	cfg := config.Config{
		Type:     api.SQLITE3,
		DB:       dbDir,
		Editor:   "vim",
		Endpoint: "",
		User:     "",
		PassEval: "",
		Pass:     "",
		Logger: config.LoggerConfig{
			Level: "DEBUG",
		},
	}
	testBufs := helpers.TestCmdBuffers{
		InBuf:  &bytes.Buffer{},
		OutBuf: &bytes.Buffer{},
		ErrBuf: &bytes.Buffer{},
	}
	headers := []string{"Question", "Answer", "Card Type", "Deck", "Due Date"}
	// TODO: Figure out why the due date is always printing today's date
	data := [][]string{
		{"A1a", "A1b", "Basic with sentences", "Default", "2021-11-22"},
		{"A2a", "A2b", "Basic with sentences", "Basic", "2021-11-22"},
	}
	args := []string{"-t", "./fixtures/templates/card-list"}
	expectedTableOut := helpers.GenerateTableOutputWithHeaders(headers, data)

	// Run command
	err := executeListCommand(t, false, &cfg, &testBufs, args, func() {
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

//func TestDefaultListQueryUsingRestCmd(t *testing.T) {
//	// Setup
//	cfg, err := config.LoadSampleConfig()
//	if err != nil {
//		os.Exit(1)
//	}
//	testBufs := helpers.TestCmdBuffers{
//		InBuf:  &bytes.Buffer{},
//		OutBuf: &bytes.Buffer{},
//		ErrBuf: &bytes.Buffer{},
//	}
//	headers := []string{"Question", "Answer", "Card Type", "Deck", "Due Date"}
//	data := [][]string{
//		{"Q2", "A2", "Basic with sentences", "Basic", "2021-11-22"},
//	}
//	expectedQuery := "deck:Basic"
//	expectedLimit := "1"
//	args := []string{"-t", "./fixtures/templates/card-list", "-q", expectedQuery, "-l", expectedLimit}
//	expectedTableOut := helpers.GenerateTableOutputWithHeaders(headers, data)
//
//	// Run command
//	err = executeListCommand(t, &cfg, &testBufs, args, func() {
//		httpmock.RegisterResponderWithQuery("GET", cfg.Endpoint+rest.CARDS_URI,
//			url.Values{"query": []string{expectedQuery}, "limit": []string{expectedLimit}},
//			httpmock.NewJsonResponderOrPanic(200, httpmock.File("./fixtures/cards/card.json")))
//	})
//	defer httpmock.DeactivateAndReset()
//
//	if err != nil {
//		t.Errorf("Could not run the cards list command: %v", err)
//	}
//
//	// Assertions
//	assert.Equal(t, "", testBufs.ErrBuf.String())
//	assert.Contains(t, expectedTableOut, testBufs.OutBuf.String())
//}
