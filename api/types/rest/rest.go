package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aerex/anki-cli/api"
	"github.com/aerex/anki-cli/internal/config"
	"github.com/aerex/anki-cli/pkg/models"
	"github.com/go-resty/resty/v2"
)

const (
  DECKS_URI = "/decks"
  COLLECTION_URI = "/collections"
)

// Structure for generic data
type DataWithMeta struct {
  Data interface {}
  Meta interface {}
}

// Structure for error response
type ErrorResponse struct {
  Code string `json:"code"`
  Message string `json:"message"`
  Source string `json:"source"`
}

type RestApi struct {
  Client *resty.Client
  Config *config.Config
}

func init() {
  api.Register(api.ApiConfig{
    Type: api.REST,
    NewApi: NewApi,
  })
}

func NewApi(config *config.Config) api.Api {
  api := &RestApi {
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

func (a RestApi) GetDecks(qs string) ([]models.Deck, error) {
  if a.Config.Endpoint != "" {
    decks := &[]models.Deck{}
    req := a.Client.R()
    req.SetResult(decks)
    if qs != "" {
      req.SetQueryParam("query", qs)
    }
    _, err := req.Get(DECKS_URI)
    if err != nil {
      return []models.Deck{}, err
    }
    return *decks, nil
  }
  // TODO: Need to move this to an error module or something
  // Same thing on line 66
  return []models.Deck{}, errors.New("could not get decks using ")
}

func (a RestApi) GetClient() *http.Client {
  return a.Client.GetClient()
}

func (a RestApi) GetStudiedStats(qs string) (models.CollectionStats, error) {
  if a.Config.Endpoint != "" {
    stats := &models.CollectionStats{}
    req := a.Client.R()
    req.SetResult(stats)
    req.SetQueryParam("include", "meta")
    if qs != "" {
      req.SetQueryParam("query", qs)
    }
    if _, err := req.Get(COLLECTION_URI); err != nil {
      return models.CollectionStats{}, err
    }
    return *stats, nil
  }
  return models.CollectionStats{}, errors.New("could not get studied stats")
}

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

func (a RestApi) CreateDeck(name string) (models.Deck, error) {
  if a.Config.Endpoint != "" {
    createdDeck := &models.Deck{}
    errorResponse := &ErrorResponse{}
    req := a.Client.R()
    req.SetResult(createdDeck)
    req.SetBody(models.Deck{Name: name})
    req.SetError(errorResponse)
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
