package repositories

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	ankisql "github.com/aerex/go-anki/api/sql"
	"github.com/aerex/go-anki/internal/utils"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/jmoiron/sqlx"
)

type ByDeckName []*models.Deck

func (d ByDeckName) Len() int           { return len(d) }
func (d ByDeckName) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d ByDeckName) Less(i, j int) bool { return d[i].Name < d[j].Name }

type deckRepo struct {
	Conn *sqlx.DB
	Tx   ankisql.TxOpts
}

type DeckRepo interface {
	Decks() (decks models.Decks, err error)
	DeckNameMap() (deckNames map[string]models.Deck, err error)
	ChildrenDeckIDs(did models.ID) (ids []models.ID, err error)
	Save(deck *models.Deck) error
	Conf(deckID models.ID) (models.DeckConfig, error)
	Confs() (deckConfs models.DeckConfigs, err error)
	Parents(deckID models.ID) (decks []models.Deck, err error)
	FixDecks(decks models.Decks, usn int) error
	DeckWithParents(deckID models.ID) ([]models.Deck, error)
}

func NewDeckRepository(conn *sqlx.DB) DeckRepo {
	return deckRepo{
		Conn: conn,
		Tx: ankisql.TxOpts{
			DB: conn,
		},
	}
}

func (d deckRepo) ChildrenDeckIDs(did models.ID) (ids []models.ID, err error) {
	decks, err := d.Decks()

	deck, exists := decks[did]
	if !exists {
		err = fmt.Errorf("deck %v does not exist", did)
		return
	}

	for _, dk := range decks {
		if strings.HasPrefix(dk.Name, fmt.Sprintf("%s:", deck.Name)) {
			ids = append(ids, dk.ID)
		}
	}
	return
}

// Save creates or updates a deck in a collection
func (d deckRepo) Save(deck *models.Deck) error {
	return ankisql.Tx(d.Tx, func(tx *sqlx.Tx) error {
		decks, err := d.Decks()
		if err != nil {
			return err
		}
		decks[deck.ID] = deck
		query := "UPDATE col SET decks = ?"
		blob, err := json.Marshal(decks)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(query, blob); err != nil {
			return err
		}
		return nil
	})
}

// Decks will retrieve the decks from col
func (d deckRepo) Decks() (decks models.Decks, err error) {
	query := `SELECT decks FROM col LIMIT 1`
	if err = d.Conn.QueryRowx(query).Scan(&decks); err != nil {
		return
	}
	return
}

func (d deckRepo) Conf(deckID models.ID) (deckConf models.DeckConfig, err error) {
	var deckConfs models.DeckConfigs
	deckConfs, err = d.Confs()

	deckConf = *deckConfs[deckID]

	return
}

func (d deckRepo) DeckNameMap() (deckNames map[string]models.Deck, err error) {
	var decks models.Decks
	decks, err = d.Decks()
	if err != nil {
		return
	}

	deckNames = make(map[string]models.Deck)
	for _, deck := range decks {
		deckNames[deck.Name] = *deck
	}

	return
}

func (d deckRepo) Confs() (deckConfs models.DeckConfigs, err error) {
	var col models.Collection
	query := `SELECT dconf from col`
	if err = d.Conn.Get(&col, query); err != nil {
		return
	}
	deckConfs = col.DeckConfs
	return
}

func (d deckRepo) Parents(deckID models.ID) (decks []models.Deck, err error) {
	var ds models.Decks
	ds, err = d.Decks()
	if err != nil {
		return
	}
	deck, exists := ds[deckID]
	if !exists {
		err = fmt.Errorf("could not find deck %s", string(deckID))
		return
	}

	var nameMap map[string]models.Deck
	nameMap, err = d.DeckNameMap()
	if err != nil {
		return
	}

	parentDecks := strings.Split(deck.Name, "::")
	parentDecks = parentDecks[:len(parentDecks)-1]
	var deckNames []string
	for _, parentDeck := range parentDecks {
		if len(deckNames) == 0 {
			deckNames = append(deckNames, parentDeck)
		} else {
			// append linked parent + grandparent deck names
			name := deckNames[len(deckNames)-1] + "::" + parentDeck
			deckNames = append(deckNames, name)
		}
	}

	for _, deckName := range deckNames {
		deck, exists := nameMap[deckName]
		if !exists {
			err = fmt.Errorf("could not find deck %s", string(deckName))
			return
		}
		decks = append(decks, deck)
	}
	return
}

func (d deckRepo) DeckWithParents(deckID models.ID) ([]models.Deck, error) {
	allDecks, err := d.Decks()
	if err != nil {
		return []models.Deck{}, err
	}
	deck, found := allDecks[deckID]
	if !found {
		return []models.Deck{}, fmt.Errorf("could not find deck %d", deckID)
	}
	decks, pErr := d.Parents(deckID)
	if pErr != nil {
		return []models.Deck{}, pErr
	}
	return append(decks, *deck), nil

}

func (d deckRepo) ensureParentsExist(immediateParents string, usn int) (string, error) {
	dbDecks, err := d.DeckNameMap()
	if err != nil {
		return "", err
	}
	parts := strings.Split(immediateParents, "::")
	var s string
	for _, p := range parts {
		if s != "" {
			s += p
		} else {
			s += "::" + p
		}
		// check if deck exists
		deck, exists := dbDecks[s]
		if exists {
			s = deck.Name
		} else {
			newDeckID := models.ID(time.Now().Unix())
			if err := d.Save(&models.Deck{Name: s, ID: newDeckID}); err != nil {
				return "", err
			}
		}
	}
	return s + "::" + parts[len(parts)-1], nil
}

func (d deckRepo) FixDecks(decks models.Decks, usn int) error {
	return ankisql.Tx(d.Tx, func(tx *sqlx.Tx) error {
		deckNames := []string{}
		var t models.UnixTime
		for _, deck := range decks {
			updateDeck := models.Deck{}
			utils.Clone(updateDeck, deck)
			// ensure deck names are unique
			if slices.Contains(deckNames, updateDeck.Name) {
				// TODO: log "fix duplicate deck names" deck.Name
				updateDeck.Name += fmt.Sprint(time.Now().Unix())
				t = models.UnixTime(time.Now().Unix())
				updateDeck.Mod = &t
				updateDeck.USN = usn
			}
			// ensure no sections are blank
			if utils.MissingParents(updateDeck.Name) {
				// TODO: log fix deck with missing sections deck.Name
				updateDeck.Name += fmt.Sprintf("recovered%d", time.Now().Unix())
				updateDeck.Mod = &t
				updateDeck.USN = usn
			}
			// immediate parent must exist
			// TODO: Finish method
			if strings.Contains(updateDeck.Name, "::") {
				// decks and subdecks (eg. School::English::Grammar)
				deckParts := strings.Split(updateDeck.Name, "::")
				// immediate decks and subdecks (eg. School::English [immediate] School::English::Grammar)
				imDeckParts := deckParts[:len(deckParts)-1]
				immediateParent := strings.Join(imDeckParts, "::")
				if !slices.Contains(deckNames, immediateParent) {
					// TODO: log fix deck with missing parent deck.Name
					dbName, err := d.ensureParentsExist(updateDeck.Name, usn)
					if err != nil {
						return err
					}
					updateDeck.Name = dbName
					deckNames = append(deckNames, immediateParent)
				}
			}

			deckNames = append(deckNames, updateDeck.Name)
		}
		return nil
	})

	return nil
}
