package api

import (
	"net/http"

	"github.com/aerex/anki-cli/internal/config"
)
const (
  PLAIN = "PLAIN"
  CUSTOM = "CUSTOM"
  REST = "REST"
)
// Method definitions for interacting with the anki api
type Api interface {

  // Get list of decks from a collection and filter list if query string is provided
  // Support simple search and regex expression 
  // See https://docs.ankiweb.net/searching.html#simple-searches
  // See https://docs.ankiweb.net/searching.html#regular-expressions
  GetDecks(qs string) (interface{}, error)
  // Get http client used in api. Useful for mocking http client in test
  GetClient() *http.Client
}

type ApiConfig struct {
  Type string
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

