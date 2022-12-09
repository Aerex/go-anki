package api

import (
	"net/http"

	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/pkg/models"
	fanki "github.com/flimzy/anki"
)

const (
	PLAIN = "PLAIN"
	REST  = "REST"
	DB    = "DB"
)

// Method definitions for interacting with the anki api
type Api interface {

	// Get list of decks from a collection and filter list if query string is provided
	// Support simple search and regex expression
	// See https://docs.ankiweb.net/searching.html#simple-searches
	// See https://docs.ankiweb.net/searching.html#regular-expressions
	GetDecks(qs string) ([]fanki.Deck, error)
	// Get http client used in api. Useful for mocking http client in test
	GetClient() *http.Client
	// Get the number of cards studied and the amount of time studied (in seconds) for a collectionin seo
	// Result can be filter by the provided query string. See GetDecks for more example usage
	GetStudiedStats(filter string) (models.CollectionStats, error)
	// Rename the deck using its ID or name
	RenameDeck(nameOrId string, newName string) (fanki.Deck, error)
	// Create a deck
	CreateDeck(name string, deckType string) (models.Deck, error)
	// Get multiple cards
	GetCards(qs string, limit int) ([]models.Card, error)
	// Get a deck study option
	GetDeckConfig(name string) (models.DeckConfig, error)
	// Get multiple deck study options
	GetAllDeckConfigs() (deckConfigs models.DeckConfigs, err error)
	// Get a card model
	GetNoteType(name string) (models.NoteType, error)
	// Get one or more card models
	// TODO: Need to remove later
	GetModels(name string) (models.NoteTypes, error)
	// Update a deck configuration
	UpdateDeckConfig(config models.DeckConfig, id string) (models.DeckConfig, error)
	// Create a card for a deck given the fields and the model
	CreateCard(note models.Note, model models.NoteType, deckName string) (models.Card, error)
}

type ApiConfig struct {
	Type   string
	NewApi func(config *config.Config) Api
}

// Called to create or find the api client based on the configured backend
func NewApi(config *config.Config) Api {
	api := ApiConfigs[config.Type]

	return api.NewApi(config)
}

var ApiConfigs = make(map[string]ApiConfig)

func Register(apiConfig ApiConfig) {
	ApiConfigs[apiConfig.Type] = apiConfig
}
