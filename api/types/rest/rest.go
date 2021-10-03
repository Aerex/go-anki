package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aerex/anki-cli/api"
	"github.com/aerex/anki-cli/internal/config"
	"github.com/aerex/anki-cli/pkg/models"
	"github.com/go-resty/resty/v2"
)

const (
  DECKS_URI = "/decks"
)
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

func (a RestApi) GetDecks(qs string) (interface{}, error) {
  if a.Config.Endpoint != "" {
    decks := &[]models.Deck{}
    req := a.Client.R()
    req.SetResult(decks)
    if qs != "" {
     req.SetQueryParam("query", qs)
    }
    _, err := req.Get(DECKS_URI)
    if err != nil {
      return nil, err
    }
    return decks, nil
  } 
  return nil, fmt.Errorf("Could not get decks using " + strings.ToLower(api.REST) + " backend type")
}

func (a RestApi) GetClient() *http.Client {
  return a.Client.GetClient()
}
