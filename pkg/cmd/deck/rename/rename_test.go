package rename

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/aerex/go-anki/api"
	"github.com/aerex/go-anki/api/rest"
	_ "github.com/aerex/go-anki/api/types"
	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/io"
	"github.com/aerex/go-anki/pkg/template"
	helpers "github.com/aerex/go-anki/tests"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func executeRenameCommand(t *testing.T, cfg *config.Config, buffers *helpers.TestCmdBuffers, args []string, mockHttp func()) error {

	anki := &anki.Anki{
		API:       api.NewApi(cfg, nil),
		Config:    cfg,
		IO:        io.NewTestIO(buffers.InBuf, buffers.OutBuf, buffers.ErrBuf, nil),
		Templates: template.NewTemplate(cfg),
	}

	client := anki.API.GetClient()
	httpmock.ActivateNonDefault(client)

	mockHttp()

	// Run deck cmd
	cmd := NewRenameCmd(anki, nil)
	if len(args) != 0 {
		cmd.SetArgs(args)
	}

	cmd.SetIn(buffers.InBuf)
	cmd.SetOut(buffers.OutBuf)
	cmd.SetErr(buffers.ErrBuf)

	err := cmd.Execute()

	return err
}

func TestRenameDeckUsingRest(t *testing.T) {
	// Setup
	oldName := "default"
	newName := "Default"
	var cfg config.Config
	if err := config.Load("../fixtures/", &cfg, &io.IO{ExecContext: exec.Command}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	variations := []struct {
		name        string
		args        []string
		expectedOut string
	}{
		{
			name:        "default",
			args:        []string{oldName, newName},
			expectedOut: fmt.Sprintf("Renamed deck to %s", newName),
		},
		{
			name:        "with q flag",
			args:        []string{oldName, newName, "-q"},
			expectedOut: "",
		},
		{
			name:        "with quiet flag",
			args:        []string{oldName, newName, "--quiet"},
			expectedOut: "",
		},
	}

	var testBufs helpers.TestCmdBuffers
	for _, variation := range variations {
		t.Run(variation.name, func(t *testing.T) {
			testBufs = helpers.TestCmdBuffers{
				InBuf:  &bytes.Buffer{},
				OutBuf: &bytes.Buffer{},
				ErrBuf: &bytes.Buffer{},
			}
			// Run command
			expectedUrl := fmt.Sprintf("%s%s/%s", cfg.Endpoint, rest.DECKS_URI, oldName)
			err := executeRenameCommand(t, &cfg, &testBufs, variation.args, func() {
				httpmock.RegisterResponder("PATCH", expectedUrl,
					httpmock.NewJsonResponderOrPanic(200, httpmock.File("../fixtures/decks/updated_deck.json")))
			})
			defer httpmock.DeactivateAndReset()
			if err != nil {
				t.Errorf("Could not run the deck rename command: %v", err)
			}
			httpmock.GetTotalCallCount()
			// Assertions
			assert.Equal(t, "", testBufs.ErrBuf.String())
			assert.Equal(t, variation.expectedOut, testBufs.OutBuf.String())
			assert.Equal(t, 1, httpmock.GetCallCountInfo()[fmt.Sprintf("PATCH %s", expectedUrl)])
		})
	}
}

func TestRenameInvalidDeckUsingRest(t *testing.T) {
	// Setup
	var cfg config.Config
	if err := config.Load("../fixtures/", &cfg, &io.IO{ExecContext: exec.Command}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	testBufs := helpers.TestCmdBuffers{
		InBuf:  &bytes.Buffer{},
		OutBuf: &bytes.Buffer{},
		ErrBuf: &bytes.Buffer{},
	}
	oldName := "default"
	newName := "Default"
	expectedUrl := fmt.Sprintf("%s%s/%s", cfg.Endpoint, rest.DECKS_URI, oldName)

	httpmock.GetTotalCallCount()

	// Run command
	url := fmt.Sprintf("%s%s/%s", cfg.Endpoint, rest.DECKS_URI, oldName)
	executeRenameCommand(t, &cfg, &testBufs, []string{oldName, newName}, func() {
		httpmock.RegisterResponder("PATCH", url,
			httpmock.NewJsonResponderOrPanic(404,
				httpmock.File("../fixtures/decks/not_found.json"),
			))
	})
	defer httpmock.DeactivateAndReset()

	// Assertions
	assert.Contains(t, testBufs.ErrBuf.String(), fmt.Sprintf("Deck %s could not be found", oldName))
	assert.Contains(t, "", testBufs.OutBuf.String())
	assert.Equal(t, 1, httpmock.GetCallCountInfo()[fmt.Sprintf("PATCH %s", expectedUrl)])
}
