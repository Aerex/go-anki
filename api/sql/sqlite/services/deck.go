package services

import (
	"sort"
	"time"

	"golang.org/x/exp/maps"

	repos "github.com/aerex/go-anki/api/sql/sqlite/repositories"
	"github.com/aerex/go-anki/pkg/models"
)

type DeckService struct {
	deckRepo repos.DeckRepo
	colRepo  repos.ColRepo
}

func NewDeckService(d repos.DeckRepo, c repos.ColRepo) DeckService {
	return DeckService{
		deckRepo: d,
		colRepo:  c,
	}
}

func (d *DeckService) fetchNewId(decks models.Decks) (id time.Time) {
	for {
		id = time.Now()
		if decks[models.ID(id.Unix())] == nil {
			break
		}
	}
	return
}

func (d *DeckService) List() (decks []*models.Deck, err error) {
	var decksMap models.Decks
	decksMap, err = d.deckRepo.Decks()
	if err != nil {
		return
	}
	decks = maps.Values(decksMap)
	sort.Sort(repos.ByDeckName(decks))
	return
}

func (d *DeckService) Create(deck *models.Deck, deckType string) (createdDeck models.Deck, err error) {
	var decks models.Decks
	decks, err = d.deckRepo.Decks()
	id := d.fetchNewId(decks)
	deck.ID = models.ID(id.Unix())
	decks, err = d.Save(deck)
	if err != nil {
		return
	}
	createdDeck = *decks[deck.ID]

	return
}

func (d *DeckService) Confs() (models.DeckConfigs, error) {
	confs, err := d.deckRepo.Confs()
	if err != nil {
		return models.DeckConfigs{}, err
	}
	return confs, nil
}

func (d *DeckService) Save(deck *models.Deck) (decks models.Decks, err error) {
	mod := models.UnixTime(time.Now().Unix())
	deck.Mod = &mod
	deck.USN, err = d.colRepo.USN()

	tx := d.deckRepo.MustCreateTrans()
	decks, err = d.deckRepo.WithTrans(tx).Create(deck)
	return
}
