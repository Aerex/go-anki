package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aerex/go-anki/api"
	"github.com/aerex/go-anki/internal/config"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/go-resty/resty/v2"
	"github.com/op/go-logging"
)

const (
	DECKS_URI        = "/decks"
	COLLECTION_URI   = "/collections"
	DECK_CONFIGS_URI = "/deckOptions"
	CARDS_URI        = "/cards"
)

var logger = logging.MustGetLogger("ankicli")

// Structure for generic data
type DataWithMeta struct {
	Data interface{}
	Meta interface{}
}

// Structure for error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Source  string `json:"source"`
}

type RestApi struct {
	Client *resty.Client
	Config *config.Config
}

func init() {
	api.Register(api.ApiConfig{
		Type:   api.REST,
		NewApi: NewApi,
	})
}

func NewApi(config *config.Config) api.Api {
	api := &RestApi{
		Client: resty.New(),
		Config: config,
	}
	api.Client.SetBasicAuth(config.User, config.Pass)
	api.Client.SetHeader("Content-Type", "application/json")
	api.Client.SetHostURL(config.Endpoint)
	return api
}

// TODO: Replace dup code with this once all tests works
//func sendRequest(model interface{}, params map[string]string, url string) {
//  decks := &[]models.Deck{}
//  req := a.Client.R()
//  req.SetResult(decks)
//  if qs != "" {
//    req.
//    req.SetQueryParams(params)
//  }
//  return req.Get(url)
//}

func (a RestApi) Decks(qs string, includeStats bool) (models.Decks, error) {
	if a.Config.Endpoint != "" {
		decks := &models.Decks{}
		req := a.Client.R()
		req.SetResult(decks)
		if qs != "" {
			req.SetQueryParam("query", qs)
		}
		_, err := req.Get(DECKS_URI)
		if err != nil {
			return models.Decks{}, err
		}
		return *decks, nil
	}
	// TODO: Need to move this to an error module or something
	// Same thing on line 66
	return models.Decks{}, errors.New("could not get decks")
}

func (a RestApi) GetClient() *http.Client {
	return a.Client.GetClient()
}

// TODO: Reimplemenat to return stats based on stats class in AnkiDroid or Anki Desktop
//func (a RestApi) GetStudiedStats(qs string) (models.CollectionStats, error) {
//	if a.Config.Endpoint != "" {
//		stats := &models.CollectionStats{}
//		req := a.Client.R()
//		req.SetResult(stats)
//		req.SetQueryParam("include", "meta")
//		if qs != "" {
//			req.SetQueryParam("query", qs)
//		}
//		if _, err := req.Get(COLLECTION_URI); err != nil {
//			return models.CollectionStats{}, err
//		}
//		return *stats, nil
//	}
//	return models.CollectionStats{}, errors.New("could not get studied stats")
//}

func (a RestApi) RenameDeck(nameOrId, newName string) (models.Deck, error) {
	if a.Config.Endpoint != "" {
		updatedDeck := &models.Deck{}
		errorResponse := &ErrorResponse{}
		req := a.Client.R()
		req.SetResult(updatedDeck)
		req.SetBody(models.Deck{Name: newName})
		req.SetError(errorResponse)
		req.SetPathParam("deckNameorId", nameOrId)

		resp, err := req.Patch(fmt.Sprintf("%s/{deckNameorId}", DECKS_URI))
		if err != nil {
			return models.Deck{}, err
		}
		if resp.IsError() {
			return models.Deck{}, errors.New(errorResponse.Message)
		}
		return *updatedDeck, nil
	}

	return models.Deck{}, errors.New("could not rename deck")
}

func (a RestApi) CreateDeck(name string, deckType string) (models.Deck, error) {
	if a.Config.Endpoint != "" {
		createdDeck := &models.Deck{}
		errorResponse := &ErrorResponse{}
		req := a.Client.R()
		req.SetResult(createdDeck)
		req.SetBody(models.Deck{Name: name})
		req.SetError(errorResponse)
		// FIXME: Should not be using a path param in a post here
		req.SetPathParam("name", name)

		resp, err := req.Post(DECKS_URI)
		if err != nil {
			return models.Deck{}, err
		}
		if resp.IsError() {
			return models.Deck{}, errors.New(errorResponse.Message)
		}
		return *createdDeck, nil
	}

	return models.Deck{}, errors.New("could not create deck")
}

func (a RestApi) GetAllDeckConfigs() (models.DeckConfigs, error) {
	panic("unimplemented")
}

func (a RestApi) GetNoteType(name string) (models.NoteType, error) {
	panic("unimplemented")
}

func (a RestApi) GetDeckConfig(nameOrId string) (models.DeckConfig, error) {
	if a.Config.Endpoint != "" {
		deckOptions := &models.DeckConfig{}
		errorResponse := &ErrorResponse{}
		req := a.Client.R()
		req.SetResult(deckOptions)
		req.SetError(errorResponse)
		req.SetPathParam("nameOrId", nameOrId)

		resp, err := req.Get(fmt.Sprintf("%s/{nameOrId}", DECK_CONFIGS_URI))
		if err != nil {
			return models.DeckConfig{}, err
		}
		if resp.IsError() {
			return models.DeckConfig{}, errors.New(errorResponse.Message)
		}
		return *deckOptions, nil
	}

	return models.DeckConfig{}, errors.New("could not get deck configs")
}

func (a RestApi) UpdateDeckConfig(deckConfig models.DeckConfig, id string) (models.DeckConfig, error) {
	if a.Config.Endpoint != "" {
		updatedDeckConfig := &models.DeckConfig{}
		errorResponse := &ErrorResponse{}
		req := a.Client.R()
		req.SetResult(updatedDeckConfig)
		req.SetBody(deckConfig)
		req.SetError(errorResponse)
		req.SetPathParam("id", id)

		resp, err := req.Patch(fmt.Sprintf("%s%s", DECK_CONFIGS_URI, id))
		if err != nil {
			return models.DeckConfig{}, err
		}
		if resp.IsError() {
			return models.DeckConfig{}, errors.New(errorResponse.Message)
		}
		return *updatedDeckConfig, nil
	}

	return models.DeckConfig{}, errors.New("could not update deck config")
}

func (a RestApi) GetCards(qs string, limit int) ([]models.Card, error) {
	if a.Config.Endpoint != "" {
		cards := &[]models.Card{}
		req := a.Client.R()
		req.SetResult(cards)
		if qs != "" {
			req.SetQueryParam("query", qs)
		}
		if limit != -1 {
			req.SetQueryParam("limit", strconv.Itoa(limit))
		}
		resp, err := req.Get(CARDS_URI)
		if err != nil || resp.IsError() {
			return []models.Card{}, err
		}
		return *cards, nil
	}
	return []models.Card{}, errors.New("could not get cards")
}

func (a RestApi) CreateCard(note models.Note, mdl models.NoteType, deckName string) (models.Card, error) {
	if a.Config.Endpoint != "" {
		createdCard := models.Card{
			Note: models.Note{
				Model: mdl,
			},
			Deck: models.Deck{
				Name: deckName,
			},
		}
		req := a.Client.R()
		req.SetResult(createdCard)
		req.SetBody(&createdCard)
		_, err := req.Post(fmt.Sprintf("%s/%s/cards", DECKS_URI, createdCard.Deck.Name))
		if err != nil {
			return models.Card{}, err
		}
		return createdCard, nil
	}
	return models.Card{}, errors.New("could not create card")
}

func (a RestApi) GetModels(name string) (models.NoteTypes, error) {
	if a.Config.Endpoint != "" {
		mdls := &models.NoteTypes{}
		req := a.Client.R()
		req.SetResult(mdls)
		if name != "" {
			req.SetQueryParam("name", name)
		}
		resp, err := req.Get(fmt.Sprintf("%s/models", COLLECTION_URI))
		if err != nil || resp.IsError() {
			return models.NoteTypes{}, err
		}
		return *mdls, nil
	}
	return models.NoteTypes{}, errors.New("could not find note type")
}
