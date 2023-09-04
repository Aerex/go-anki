package repositories

import (
	"database/sql"

	"github.com/aerex/go-anki/pkg/models"

	"github.com/jmoiron/sqlx"
)

type cardRepo struct {
	Conn *sqlx.DB
	Tx   *sqlx.Tx
}

type CardRepo interface {
	List(cls string, args []string) (cards []models.Card, err error)
	Get(ID string, query string, data interface{}) error
	Exists(cardId int64) (err error, exists bool)
	Create(card models.Card) (err error)
	WithTrans(trans interface{}) CardRepo
	MustCreateTrans() *sqlx.Tx
	CardsDueForDeck(deckId int64, due int64, limit int) (count int64, err error)
	CardsLearnedForDeck(deckId int64, due int64, limit int) (count int64, err error)
	CardsNewForDeck(deckId int64, limit int) (count int64, err error)
}

func NewCardRepository(conn *sqlx.DB) CardRepo {
	return cardRepo{
		Conn: conn,
	}
}

func (c cardRepo) List(cls string, args []string) ([]models.Card, error) {
	// FIXME: Need to figure out why we need to use an []interface slice here
	var cards []models.Card
	var iargs []interface{} = make([]interface{}, len(args))
	for i, d := range args {
		iargs[i] = d
	}
	baseQuery := `SELECT cards.*, note.flds "note.flds", note.mid "note.mid" FROM cards
    JOIN notes note ON note.id = cards.nid`
	var rows *sqlx.Rows
	if cls != "" {
		query, arglist, err := sqlx.In(baseQuery+" WHERE "+cls, iargs...)
		if err != nil {
			return cards, err
		}
		query = c.Conn.Rebind(query)
		rows, err = c.Conn.Queryx(query, arglist...)
		if err != nil {
			return cards, err
		}
	} else {
		var err error
		rows, err = c.Conn.Queryx(baseQuery)
		if err != nil {
			return cards, err
		}
	}

	for rows.Next() {
		var card models.Card
		rows.StructScan(&card)
		cards = append(cards, card)
	}
	err := rows.Err()

	return cards, err
}

func (c cardRepo) Get(ID string, query string, data interface{}) error {
	err := c.Conn.QueryRowx(query, ID).StructScan(&data)
	if err != nil {
		return err
	}
	return nil
}

func (c cardRepo) Exists(cardId int64) (err error, exists bool) {
	var card models.Card
	query := "SELECT id FROM cards WHERE id = ?"
	err = c.Conn.Get(&card, query, cardId)
	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		return err, false
	}
	return
}

func (c cardRepo) CardsDueForDeck(deckId int64, due int64, limit int) (count int64, err error) {
	query := `SELECT COUNT() FROM (SELECT 1 FROM cards WHERE did = ? AND queue = 2
    AND due <= ?`
	if limit != 0 {
		query = query + " LIMIT " + string(limit) + ")"
	} else {
		query = query + ")"
	}
	row := c.Conn.QueryRow(query, deckId, due)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	return
}

func (c cardRepo) CardsNewForDeck(deckId int64, limit int) (count int64, err error) {
	query := `SELECT COUNT() FROM (SELECT 1 FROM cards WHERE did = ? AND queue = 0`
	if limit != 0 {
		query = query + " LIMIT " + string(limit) + ")"
	} else {
		query = query + ")"
	}
	row := c.Conn.QueryRow(query, deckId)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	return
}

func (c cardRepo) CardsLearnedForDeck(deckId int64, due int64, limit int) (count int64, err error) {
	query := `SELECT SUM(left/1000) FROM (SELECT left FROM cards WHERE did = ? AND queue = 1 AND due < ?`
	if limit != 0 {
		query = query + " LIMIT " + string(limit) + ")"
	} else {
		query = query + ")"
	}
	row := c.Conn.QueryRow(query, deckId, due)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	return
}

func (c cardRepo) Create(card models.Card) (err error) {
	query := `INSERT OR REPLACE INTO cards (nid, did, ord, mod, usn, type, queue, due, ivl, factor, reps, lapses, left, odue, odid, flags, data) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?, "")`
	if c.Tx != nil {
		if _, err = c.Tx.Exec(query, card.NoteID, card.DeckID, card.Ord, card.Mod, card.USN, card.Type, card.Queue,
			card.Due, card.Interval, card.Factor, card.Reps, card.Lapses, card.Left, card.Odue, card.OriginalDeckID, card.Flags); err != nil {
			return
		}
		return
	}
	if _, err = c.Conn.NamedExec(query, card); err != nil {
		return
	}
	return
}

func (c cardRepo) WithTrans(trans interface{}) CardRepo {
	c.Tx = trans.(*sqlx.Tx)
	return c
}

func (c cardRepo) MustCreateTrans() *sqlx.Tx {
	return c.Conn.MustBegin()
}
