package services

import (
	"fmt"
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

func (d *DeckService) Create(deck *models.Deck) (err error) {
	decks, err := d.deckRepo.Decks()
	id := d.fetchNewId(decks)
	deck.ID = models.ID(id.Unix())
	if err := d.Save(deck); err != nil {
		return err
	}
	return nil
}

// Rename renames an existing deck in a collection.
func (d *DeckService) Rename(name, newName string) error {
	deckNameMap, err := d.deckRepo.DeckNameMap()
	if err != nil {
		return err
	}
	deck, exists := deckNameMap[name]
	if !exists {
		return fmt.Errorf("could not find Deck %s", name)
	}
	deck.Name = newName
	dconfs, err := d.deckRepo.Confs()
	if err != nil {
		return err
	}
	dconf := dconfs[deck.ID]
	dconf.Name = newName
	dconfs[deck.ID] = dconf
	return d.Save(&deck)
}

func (d *DeckService) Confs() (models.DeckConfigs, error) {
	confs, err := d.deckRepo.Confs()
	if err != nil {
		return models.DeckConfigs{}, err
	}
	return confs, nil
}

func (d *DeckService) Save(deck *models.Deck) error {
	mod := models.UnixTime(time.Now().Unix())
	deck.Mod = &mod
	var usnErr error
	deck.USN, usnErr = d.colRepo.USN(false)
	if usnErr != nil {
		return usnErr
	}
	if err := d.deckRepo.Save(deck); err != nil {
		return err
	}

	return nil
}
