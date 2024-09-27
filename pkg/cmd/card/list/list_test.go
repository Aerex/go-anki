package list

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/aerex/go-anki/api"
	"github.com/aerex/go-anki/api/apifakes"
	_ "github.com/aerex/go-anki/api/types"
	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/pkg/anki"
	"github.com/aerex/go-anki/pkg/io"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/aerex/go-anki/pkg/template"
	helpers "github.com/aerex/go-anki/tests"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

var fakeApi apifakes.FakeApi

func executeListCommand(t *testing.T, useHttp bool, cfg *config.Config, buffers *helpers.TestCmdBuffers, args []string, mockHttp func()) error {

	fakeApi = apifakes.FakeApi{}
	anki := &anki.Anki{
		API:       &fakeApi,
		Config:    cfg,
		IO:        io.NewTestIO(buffers.InBuf, buffers.OutBuf, buffers.ErrBuf, exec.Command),
		Templates: template.NewTemplate(cfg),
	}

	if useHttp {
		client := anki.API.GetClient()
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

func TestListCardsSQL(t *testing.T) {
	// Setup
	_, fileName, _, _ := runtime.Caller(0)
	loc, err := time.LoadLocation("Local")
	if err != nil {
		t.Errorf("Could not get date location: %v", err)
	}
	due, err := time.ParseInLocation("2006-01-02", "2021-11-21", loc)
	if err != nil {
		t.Errorf("Could not parse date: %v", err)
	}
	basePath := filepath.Join(filepath.Dir(fileName))
	dbDir := filepath.Join(basePath, "/fixtures/db/collection.anki2")
	cfg := config.Config{
		Type: api.DB,
		DB: config.DBConfig{
			Driver: api.SQLITE3,
			Path:   dbDir,
		},
		Dir: filepath.Join(basePath, "/fixtures/configs/sql"),
		Logger: config.LoggerConfig{
			Level: "DEBUG",
		},
	}
	cards := []models.Card{
		{
			Deck: models.Deck{
				Name: "Default",
			},
			Due: models.UnixTime(due.Unix()),
			Note: models.Note{
				Fields: []string{
					"Question",
					"Answer",
				},
				Model: models.NoteType{
					Name: "Basic",
					Tags: []string{},
					Fields: []*models.CardField{
						{
							Name:    "Front",
							Ordinal: 0,
						},
						{
							Name:    "Back",
							Ordinal: 1,
						},
					},
					SortField: 0,
					Templates: []*models.CardTemplate{
						{
							AnswerFormat:   "{{Back}}",
							QuestionFormat: "{{Front}}",
							Name:           "Card",
							Ordinal:        0,
						},
					},
					Type: models.StandardCardType,
				},
			},
		},
	}

	fakeApi = apifakes.FakeApi{}
	fakeApi.GetCardsReturns(cards, nil)
	testBufs := helpers.TestCmdBuffers{
		InBuf:  &bytes.Buffer{},
		OutBuf: &bytes.Buffer{},
		ErrBuf: &bytes.Buffer{},
	}
	anki := &anki.Anki{
		API:       &fakeApi,
		Config:    &cfg,
		IO:        io.NewTestIO(testBufs.InBuf, testBufs.OutBuf, testBufs.ErrBuf, exec.Command),
		Templates: template.NewTemplate(&cfg),
	}

	tests := []struct {
		name              string
		cliArgs           []string
		useCustomTemplate bool
		expectErr         bool
	}{
		{
			name: "No args",
		},
		{
			name:    "with limit",
			cliArgs: []string{"-l", "15"},
		},
		{
			name:      "with invalid limit",
			cliArgs:   []string{"-l", "-3"},
			expectErr: true,
		},
		{
			name:    "with query",
			cliArgs: []string{"-q", "Question"},
		},
		{
			name:              "with json template",
			cliArgs:           []string{"-t", "json"},
			useCustomTemplate: true,
		},
	}
	for _, tt := range tests {
		cmd := NewListCmd(anki)
		if !tt.useCustomTemplate {
			tt.cliArgs = append(tt.cliArgs, "-t")
			tt.cliArgs = append(tt.cliArgs, "./fixtures/templates/card-list")
		}
		if len(tt.cliArgs) != 0 {
			cmd.SetArgs(tt.cliArgs)
		}
		cmd.SetIn(testBufs.InBuf)
		cmd.SetOut(testBufs.OutBuf)
		cmd.SetErr(testBufs.ErrBuf)

		err := cmd.Execute()
		if tt.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
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
