package api

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"net/http"

	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/rs/zerolog"
)

const (
	PLAIN   = "PLAIN"
	REST    = "REST"
	SQLITE3 = "SQLITE3"
	DB      = "db"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Api
// Method definitions for interacting with the anki api
type Api interface {
	// DeckStudyStats provides stats for the number of new/reviewed/learning cards per deck
	DeckStudyStats() (stats map[models.ID]models.DeckStudyStats, err error)
	// Get list of decks from a collection and filter list if query string is provided
	// optionally include stats
	// Support simple search and regex expression
	// See https://docs.ankiweb.net/searching.html#simple-searches
	// See https://docs.ankiweb.net/searching.html#regular-expressions
	Decks(qs string) ([]*models.Deck, error)
	// Get http client used in api. Useful for mocking http client in test
	GetClient() *http.Client
	// TODO: Reimplemented StudiedStats method based on Stats class in AnkiDroid or Anki Desktop
	// Get the number of cards studied and the amount of time studied (in seconds) for a collectionin seo
	// Result can be filter by the provided query string. See GetDecks for more example usage
	//GetStudiedStats(filter string) (models.CollectionStats, error)
	// Rename the deck using its ID or name
	RenameDeck(nameOrId string, newName string) error
	// Create a deck
	CreateDeck(name string) error
	// Get multiple cards. To return all cards pass -1 for the limit
	Cards(qs string, limit int) ([]models.Card, error)
	// Get a deck study option
	GetDeckConfig(name string) (models.DeckConfig, error)
	// Get multiple deck study options
	GetAllDeckConfigs() (deckConfigs models.DeckConfigs, err error)
	// Get a card model
	NoteType(name string) (models.NoteType, error)
	// Get one or more card models
	// TODO: Need to remove later
	NoteTypes() (models.NoteTypes, error)
	// Update a deck configuration
	UpdateDeckConfig(config models.DeckConfig, id string) (models.DeckConfig, error)
	// Create a card for a deck given the fields and the model
	CreateCard(note models.Note, model models.NoteType, deckName string) (models.Card, error)
	// Tags returns a list of tags cached in the collection
	Tags() ([]string, error)
}

type ApiConfig struct {
	Type   string
	NewApi func(config *config.Config, log *zerolog.Logger) Api
}

// Called to create or find the api client based on the configured backend
func NewApi(config *config.Config, log *zerolog.Logger) Api {
	api := ApiConfigs[config.General.Type]

	return api.NewApi(config, log)
}

var ApiConfigs = make(map[string]ApiConfig)

func Register(apiConfig ApiConfig) {
	ApiConfigs[apiConfig.Type] = apiConfig
}
