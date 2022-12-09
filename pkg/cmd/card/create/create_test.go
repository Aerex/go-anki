package create

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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
)

func fakeExecEditorSuccess(command string, args ...string) *exec.Cmd {
	cmdString := []string{"-test.run=TestProcessSuccess", "--", command}
	cmdString = append(cmdString, args...)
	cmd := exec.Command(os.Args[0], cmdString...)
	cmd.Env = []string{"SKIP_TEST_MOCK=1"}
	return cmd
}

// TestShellProcessSuccess is a method that is called as a substitute for a shell command,
// the SKIP_TEST_MOCK flag ensures that if it is called as part of the test suite, it is
// skipped.
func TestProcessSuccess(t *testing.T) {
	if os.Getenv("SKIP_TEST_MOCK") != "1" {
		return
	}
	os.Exit(0)
}

func executeCreateCommand(t *testing.T, cfg *config.Config, buffers *helpers.TestCmdBuffers, args []string, mockHttp func(), mockExecCtx io.ExecContext) error {

	anki := &anki.Anki{
		Api:       api.NewApi(cfg),
		Config:    cfg,
		IO:        io.NewTestIO(buffers.InBuf, buffers.OutBuf, buffers.ErrBuf, mockExecCtx),
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

func TestCreateCardUsingRest(t *testing.T) {
	// Setup
	deckName := "Default"
	cfg, err := config.LoadSampleConfig()
	if err != nil {
		os.Exit(1)
	}
	testBufs := helpers.TestCmdBuffers{
		InBuf:  &bytes.Buffer{},
		OutBuf: &bytes.Buffer{},
		ErrBuf: &bytes.Buffer{},
	}
	expectedOut := "Created new card"
	args := []string{"--field", "A2b - Front", "--field", "A2b - Back", "--deck", deckName, "--no-edit"}
	// Run command
	expectedCreateCardUrl := fmt.Sprintf("%s%s/%s/cards", cfg.Endpoint, rest.DECKS_URI, deckName)
	expectedGetModelUrl := fmt.Sprintf("%s%s/models", cfg.Endpoint, rest.COLLECTION_URI)
	actualCard := &models.Card{}
	err = executeCreateCommand(t, &cfg, &testBufs, args, func() {
		httpmock.RegisterResponder("GET", expectedGetModelUrl,
			httpmock.NewJsonResponderOrPanic(200, httpmock.File("./fixtures/model.json")))

		httpmock.RegisterResponder("POST", expectedCreateCardUrl,
			func(req *http.Request) (*http.Response, error) {
				if err := json.NewDecoder(req.Body).Decode(&actualCard); err != nil {
					return httpmock.NewStringResponse(400, ""), nil
				}
				return httpmock.NewJsonResponse(200, httpmock.File("./fixtures/card.json"))
			})
	}, nil)
	defer httpmock.DeactivateAndReset()
	if err != nil {
		t.Errorf("Could not run the create card command: %v", err)
	}

	// Assertions
	httpmock.GetTotalCallCount()
	assert.Equal(t, expectedOut, testBufs.OutBuf.String())
	assert.Equal(t, 1, httpmock.GetCallCountInfo()[fmt.Sprintf("GET %s", expectedGetModelUrl)])
	assert.Equal(t, 1, httpmock.GetCallCountInfo()[fmt.Sprintf("POST %s", expectedCreateCardUrl)])
}

// TODO: Refactor and combine with previous test once pass
func TestCreateCardWithEditorWithRest(t *testing.T) {
	// FIXME:: Figure out how to mock io channel in test
	//t.SkipNow()
	// Setup
	deckName := "Default"
	cfg, err := config.LoadSampleConfig()
	if err != nil {
		os.Exit(1)
	}
	testBufs := helpers.TestCmdBuffers{
		InBuf:  &bytes.Buffer{},
		OutBuf: &bytes.Buffer{},
		ErrBuf: &bytes.Buffer{},
	}
	expectedOut := "Created new card"
	args := []string{"--field", "A2b - Front", "--field", "A2b - Back", "--deck", deckName}
	// Run command
	expectedCreateCardUrl := fmt.Sprintf("%s%s/%s/cards", cfg.Endpoint, rest.DECKS_URI, deckName)
	expectedGetModelUrl := fmt.Sprintf("%s%s/models", cfg.Endpoint, rest.COLLECTION_URI)
	actualCard := &models.Card{}
	err = executeCreateCommand(t, &cfg, &testBufs, args, func() {
		httpmock.RegisterResponder("GET", expectedGetModelUrl,
			httpmock.NewJsonResponderOrPanic(200, httpmock.File("./fixtures/model.json")))

		httpmock.RegisterResponder("POST", expectedCreateCardUrl,
			func(req *http.Request) (*http.Response, error) {
				if err := json.NewDecoder(req.Body).Decode(&actualCard); err != nil {
					return httpmock.NewStringResponse(400, ""), nil
				}
				return httpmock.NewJsonResponse(200, httpmock.File("./fixtures/card.json"))
			})
	}, fakeExecEditorSuccess)
	defer httpmock.DeactivateAndReset()
	if err != nil {
		t.Errorf("Could not run the create card command: %v", err)
	}

	// Assertions
	httpmock.GetTotalCallCount()
	assert.Equal(t, expectedOut, testBufs.OutBuf.String())
	assert.Equal(t, 1, httpmock.GetCallCountInfo()[fmt.Sprintf("GET %s", expectedGetModelUrl)])
	assert.Equal(t, 1, httpmock.GetCallCountInfo()[fmt.Sprintf("POST %s", expectedCreateCardUrl)])
}
