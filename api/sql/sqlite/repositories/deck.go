package repositories

import (
	"fmt"
	"strings"

	"github.com/aerex/go-anki/pkg/models"
	"github.com/jmoiron/sqlx"
)

type deckRepo struct {
	Conn *sqlx.DB
	Tx   *sqlx.Tx
}

type DeckRepo interface {
	Decks() (decks models.Decks, err error)
	DeckIDs(did models.ID) (ids []models.ID, err error)
	Create(deck *models.Deck) (decks models.Decks, err error)
	Conf(deckId models.ID) (models.DeckConfig, error)
	Confs() (deckConfs models.DeckConfigs, err error)
	WithTrans(trans interface{}) DeckRepo
	MustCreateTrans() *sqlx.Tx
}

func NewDeckRepository(conn *sqlx.DB) DeckRepo {
	return deckRepo{
		Conn: conn,
	}
}

func (d deckRepo) DeckIDs(did models.ID) (ids []models.ID, err error) {
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

// Create creates a new deck to a collection
func (d deckRepo) Create(deck *models.Deck) (decks models.Decks, err error) {
	decks, err = d.Decks()
	decks[deck.ID] = deck
	query := `INSERT INTO col (decks) VALUES (?)`
	// TODO: Figure out how to abstract using either db or tx
	if d.Tx != nil {
		if _, err = d.Tx.Exec(query, decks); err != nil {
			return
		}
		return
	}
	if _, err = d.Conn.Exec(query, decks); err != nil {
		return
	}
	return
}

// Decks will retrieve the decks from col
func (d deckRepo) Decks() (decks models.Decks, err error) {
	query := `SELECT decks FROM col LIMIT 1`
	if err = d.Conn.QueryRowx(query).Scan(&decks); err != nil {
		return
	}
	return
}

func (d deckRepo) Conf(deckId models.ID) (deckConf models.DeckConfig, err error) {
	var deckConfs models.DeckConfigs
	deckConfs, err = d.Confs()

	deckConf = *deckConfs[deckId]

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

func (d deckRepo) WithTrans(trans interface{}) DeckRepo {
	d.Tx = trans.(*sqlx.Tx)
	return d
}

func (d deckRepo) MustCreateTrans() *sqlx.Tx {
	return d.Conn.MustBegin()
}

//func Migrate() (err error) {
//	if err = migrate(); err != nil {
//		return
//	}
//	return
//}
