package sqlite

import (
	"net/http"
	"strings"

	repos "github.com/aerex/go-anki/api/sql/sqlite/repositories"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/rs/zerolog"
	sqldblogger "github.com/simukti/sqldb-logger"

	"github.com/aerex/go-anki/api"
	"github.com/aerex/go-anki/api/sql/sqlite/services"
	schedv2 "github.com/aerex/go-anki/api/sql/sqlite/services/sched/v2"
	"github.com/aerex/go-anki/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
)

func init() {
	api.Register(api.ApiConfig{
		Type:   api.DB,
		NewApi: NewApi,
	})
}

type SqliteApi struct {
	Config         *config.Config
	CardService    services.CardService
	ColService     services.ColService
	DeckService    services.DeckService
	SchedV2Service schedv2.SchedV2Service
}

func NewApi(config *config.Config, log *zerolog.Logger) api.Api {
	api := &SqliteApi{
		Config: config,
	}
	db := sqlx.MustConnect(strings.ToLower(config.DB.Driver), config.DB.File)

	// enable sql logging
	if config.Logger.Sql {
		db.DB = sqldblogger.OpenDriver(config.DB.File, db.Driver(), zerologadapter.New(*log))
	}
	cardRepo := repos.NewCardRepository(db)
	colRepo := repos.NewColRepository(db)
	deckRepo := repos.NewDeckRepository(db)
	noteRepo := repos.NewNoteRepository(db)
	api.CardService = services.NewCardService(cardRepo, colRepo, deckRepo, noteRepo)
	api.ColService = services.NewColService(colRepo)
	api.DeckService = services.NewDeckService(deckRepo, colRepo)
	// TODO: Figure out how to handle the server property
	// @see third parameter in NewSchedService method
	api.SchedV2Service = schedv2.NewSchedService(colRepo, cardRepo, deckRepo, true)
	return api
}

// GetClient implements api.Api
func (*SqliteApi) GetClient() *http.Client {
	panic("Expecting RestApi but got SqliteApi")
}

// GetDeckConfig implements api.Api
func (*SqliteApi) GetDeckConfig(name string) (models.DeckConfig, error) {
	panic("unimplemented")
}

// Decks implements api.Api
func (a *SqliteApi) Decks(qs string) ([]*models.Deck, error) {
	return a.DeckService.List()
}

func (a *SqliteApi) DeckStudyStats() (stats map[models.ID]models.DeckStudyStats, err error) {
	return a.SchedV2Service.DeckStudyStats()
	//switch a.Config.SchedulerVersion {
	//case 2:
	//	return a.SchedV2Service.DeckStudyStats()
	//default:
	//}
	//return
}

// GetStudiedStats implements api.Api
func (a *SqliteApi) GetStudiedStats(filter string) (models.CollectionStats, error) {
	panic("unimplemented")
}

// RenameDeck rename the deck provided
func (a SqliteApi) RenameDeck(name string, newName string) error {
	return a.DeckService.Rename(name, newName)
}

// UpdateDeckConfig implements api.Api
func (*SqliteApi) UpdateDeckConfig(config models.DeckConfig, id string) (models.DeckConfig, error) {
	panic("unimplemented")
}

func (a SqliteApi) GetAllDeckConfigs() (deckConfigs models.DeckConfigs, err error) {
	deckConfigs, err = a.DeckService.Confs()
	if err != nil {
		return
	}
	return
}

func (a SqliteApi) NoteType(name string) (noteType models.NoteType, err error) {
	noteType, err = a.ColService.GetNoteTypeByName(name)
	if err != nil {
		return
	}
	return
}

func (a SqliteApi) CreateCard(note models.Note, noteType models.NoteType, deckName string) (createdCard models.Card, err error) {
	for _, tmpl := range noteType.Templates {
		if err = a.CardService.Create(createdCard, note, noteType, *tmpl, deckName); err != nil {
			return
		}
	}
	return
}

func (a SqliteApi) Cards(qs string, limit int) (cards []models.Card, err error) {
	cards, err = a.CardService.Find(qs)
	if limit > 0 && limit < len(cards) {
		return cards[0:limit], nil
	}
	return
}

func (a SqliteApi) CreateDeck(name string) (err error) {
	if err = a.DeckService.Create(&models.Deck{Name: name}); err != nil {
		return
	}
	return
}

func (a *SqliteApi) NoteTypes() (types models.NoteTypes, err error) {
	return a.ColService.NoteTypes()
}

func (a *SqliteApi) Tags() ([]string, error) {
	return a.ColService.Tags()
}
