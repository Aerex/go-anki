package repositories

import (
	dbsql "database/sql"
	"fmt"
	"time"

	ankisql "github.com/aerex/go-anki/api/sql"

	"github.com/aerex/go-anki/pkg/models"
	"github.com/google/gapid/core/math/sint"

	"github.com/jmoiron/sqlx"
)

type cardRepo struct {
	Conn *sqlx.DB
	Tx   ankisql.TxOpts
}

var RESTORE_QUEUE_SNIPPET = fmt.Sprintf("queue = (CASE WHEN type IN (1, %d) THEN"+
	" (CASE WHEN (CASE WHEN odue THEN odue ELSE due END) > 1000000000 THEN 1 ELSE"+
	" %d END)"+
	" ELSE"+
	" type"+
	" END)", models.CardTypeRelearning, models.CardQueueRelearning)

var RESTORE_QUEUE_WHEN_EMPTYING_SNIPPET = fmt.Sprintf("queue = (CASE WHEN queue < 0 THEN queue "+
	"WHEN type IN (1, %s) THEN "+
	"(CASE WHEN (CASE WHEN odue THEN odue ELSE due END) > 1000000000 THEN 1 ELSE "+
	"%s end)", models.CardTypeRelearning, models.CardQueueRelearning) +
	" ELSE " +
	"type end)"

type CardRepo interface {
	List(cls string, args []string) (cards []models.Card, err error)
	Exists(cardID int64) (err error, exists bool)
	Create(card models.Card) (err error)
	Update(card models.Card) (err error)
	CardsDueForDeck(deckID int64, due int64, limit int) (lrnCnt int64, err error)
	CardsLearnedForDeck(deckID int64, due int64, today int64, limit int) (count int, err error)
	CardsNewForDeck(deckID models.ID, limit int) (count int, err error)
	CardsReviewForDeck(deckLimit string, reportLimit int, reviewLimit int, today int) (count int, err error)
	UnburyCards() (err error)
	BuriedCards(noteID models.ID, cardID models.ID) (cards []models.Card, err error)
	BuryCards(cardIDs []models.ID, usn int) error
	RecoverOrphans(deckLimit string) (err error)
	LearningCount(deckLimit string, lrnCutoff int64) (count int, err error)
	Revisions(deckLimit string, limit int) (count int, err error)
	NewCardsCount(deckID models.ID, limit int) (count int, err error)
	EmptyDyn(limitQuery string, usn int) error
}

func NewCardRepository(conn *sqlx.DB) CardRepo {
	return cardRepo{
		Conn: conn,
		Tx: ankisql.TxOpts{
			DB: conn,
		},
	}
}

func (c cardRepo) List(cls string, args []string) ([]models.Card, error) {
	// FIXME: Need to figure out why we need to use an []interface slice here
	var cards []models.Card
	var iargs []interface{} = make([]interface{}, len(args))
	for i, d := range args {
		iargs[i] = d
	}
	baseQuery := "SELECT c.*, note.flds \"note.flds\", note.mid \"note.mid\" FROM cards \"c\"" +
		" JOIN notes note ON note.id = c.nid"
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

func (c cardRepo) Exists(cardId int64) (err error, exists bool) {
	var card models.Card
	query := "SELECT id FROM cards WHERE id = ?"
	err = c.Conn.Get(&card, query, cardId)
	if err == dbsql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		return err, false
	}
	return
}

func (c cardRepo) CardsDueForDeck(deckId int64, due int64, limit int) (count int64, err error) {
	query := "SELECT COUNT() FROM (SELECT 1 FROM cards WHERE did = ? AND queue = 2" +
		" AND due <= ?"
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
func (c cardRepo) CardsReviewForDeck(deckLimit string, reportLimit int, reviewLimit int, today int) (count int, err error) {
	lim := sint.Min(reportLimit, reviewLimit)
	query := fmt.Sprintf("SELECT COUNT() FROM (SELECT 1 FROM cards WHERE did IN %s AND queue = %d AND due <= ? LIMIT ?)", deckLimit, models.CardTypeReview)
	row := c.Conn.QueryRow(query, today, lim)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	return
}

func (c cardRepo) CardsNewForDeck(deckID models.ID, limit int) (count int, err error) {
	query := fmt.Sprintf("SELECT COUNT() FROM (SELECT 1 FROM cards WHERE did = ?  AND queue = %d LIMIT ?)", models.CardTypeNew)
	row := c.Conn.QueryRow(query, deckID, limit)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	return
}

func (c cardRepo) CardsLearnedForDeck(deckId int64, due int64, today int64, limit int) (count int, err error) {
	query := fmt.Sprintf("SELECT COUNT() FROM (SELECT NULL FROM cards WHERE did = ? AND queue = %d AND due < ? limit ?)", models.CardTypeNew)
	row := c.Conn.QueryRow(query, deckId, due, limit)
	err = row.Scan(&count)
	if err != nil {
		return
	}

	var relearn int
	query = fmt.Sprintf("SELECT COUNT() FROM (SELECT NULL FROM cards WHERE did = ? AND queue = %d AND due <= ? limit ?)", models.CardQueueRelearning)
	row = c.Conn.QueryRow(query, deckId, today, limit)
	err = row.Scan(&relearn)
	if err != nil {
	}
	count += relearn

	return
}

// UnburyCards will unbury all buried cards in all decks."
func (c cardRepo) UnburyCards() (err error) {
	return ankisql.Tx(c.Tx, func(tx *sqlx.Tx) error {
		query := fmt.Sprintf("UPDATE cards SET "+RESTORE_QUEUE_SNIPPET+" WHERE queue in (-2, %d)", models.CardQueueBuried)
		if _, err := tx.Exec(query); err != nil {
			return err
		}
		return nil
	})
}

func (c cardRepo) RecoverOrphans(deckLimit string) (err error) {
	return ankisql.Tx(c.Tx, func(tx *sqlx.Tx) error {
		query := "UPDATE cards SET did = 1 WHERE did NOT IN " + deckLimit
		if _, err = tx.Exec(query); err != nil {
			return err
		}
		return nil
	})
}

func (c cardRepo) Create(card models.Card) (err error) {
	return ankisql.Tx(c.Tx, func(tx *sqlx.Tx) error {
		query := `INSERT OR REPLACE INTO cards (nid, did, ord, mod, usn, type, queue, due, ivl, factor, reps, lapses, left, odue, odid, flags, data) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?, "")`
		if _, err := tx.Exec(query, card.NoteID, card.DeckID, card.Ord, card.Mod, card.USN, card.Type, card.Queue,
			card.Due, card.Interval, card.Factor, card.Reps, card.Lapses, card.ReviewsLeft, card.OriginalDue, card.OriginalDeckID, card.Flags); err != nil {
			return err
		}
		return nil
	})
}

func (c cardRepo) Update(card models.Card) (err error) {
	return ankisql.Tx(c.Tx, func(tx *sqlx.Tx) error {
		query := `UPDATE cards SET mod=?, usn=?, type=?, queue=?, due=?, ivl=?, factor=?, reps=?,
    lapses=?, left=?, odue=?, odid=?, did=? where id = ?`
		if _, err := tx.Exec(query, card.Mod, card.USN, card.Type, card.Queue, card.Due, card.Interval, card.Factor,
			card.Reps, card.Lapses, card.ReviewsLeft, card.OriginalDue, card.OriginalDeckID, card.DeckID, card.ID); err != nil {
			return err
		}
		return nil
	})
}

func (c cardRepo) LearningCount(deckLimit string, lrnCutoff int64) (lrnCnt int, err error) {
	var count int
	// subday
	query := fmt.Sprintf("SELECT SUM(left/1000) FROM (SELECT left FROM cards WHERE did IN %s AND queue = %d AND due > ? LIMIT ",
		deckLimit, models.CardQueueLearning)
	row := c.Conn.QueryRow(query, lrnCutoff)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	lrnCnt += count

	// day
	query = fmt.Sprintf("SELECT COUNT() FROM cards WHERE did IN %s AND queue = %d AND due <= ?",
		deckLimit, models.CardQueueRelearning)
	today := time.Now().Unix()
	row = c.Conn.QueryRow(query, today)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	lrnCnt += count

	// previews
	query = fmt.Sprintf("SELECT COUNT() FROM cards WHERE did IN %s AND queue = %d", deckLimit, models.CardQueueReview)
	row = c.Conn.QueryRow(query)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	lrnCnt += count

	return
}

func (c cardRepo) Revisions(deckLimit string, limit int) (count int, err error) {
	query := fmt.Sprintf("SELECT COUNT() FROM "+
		"(SELECT ID FROM cards WHERE did IN %s AND queue = %d AND due <= ? limit ?)",
		deckLimit, models.CardTypeReview)
	row := c.Conn.QueryRow(query, time.Now().Unix(), limit)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	return
}

func (c cardRepo) NewCardsCount(deckID models.ID, deckLimit int) (count int, err error) {
	query := fmt.Sprintf("SELECT COUNT() FROM (SELECT 1 FROM cards WHERE did = ? AND queue = %d LIMIT ?", models.CardTypeNew)
	row := c.Conn.QueryRow(query, deckID, deckLimit)
	err = row.Scan(&count)
	if err != nil {
		return
	}
	return
}

func (c cardRepo) BuriedCards(noteID models.ID, cardID models.ID) (cards []models.Card, err error) {
	query := "SELECT id, queue FROM cards WHERE nid=? and id!=? AND " +
		"(queue=0 OR (queue=2 AND due <=?))"
	query = c.Conn.Rebind(query)
	today := time.Now().Unix()
	rows, err := c.Conn.Queryx(query, noteID, cardID, today)
	if err != nil {
		return
	}

	for rows.Next() {
		var card models.Card
		rows.StructScan(&card)
		cards = append(cards, card)
	}
	err = rows.Err()

	return
}

func (c cardRepo) EmptyDyn(limitQuery string, usn int) error {
	return ankisql.Tx(c.Tx, func(tx *sqlx.Tx) error {
		query := fmt.Sprintf("UPDATE cards SET did = odid, %s, due = (CASE WHEN odue > 0 THEN odue ELSE due END), "+
			"odue = 0, odid = 0, usn = ? WHERE %s", RESTORE_QUEUE_WHEN_EMPTYING_SNIPPET, limitQuery)
		query = c.Conn.Rebind(query)
		if _, err := tx.Exec(query, usn); err != nil {
			return err
		}
		return nil
	})
}

func (c cardRepo) BuryCards(cardIDs []models.ID, usn int) error {
	return ankisql.Tx(c.Tx, func(tx *sqlx.Tx) error {
		queue := fmt.Sprintf("manual AND %s OR %s", models.CardQueueBuried, models.CardQueueSBuried)
		query := "UPDATE cards SET queue = ?, mod = ?, usn = ? WHERE id IN " +
			ankisql.InClauseFromIDs(cardIDs)

		query = c.Conn.Rebind(query)
		if _, err := tx.Exec(query, queue, time.Now().Unix(), usn); err != nil {
			return err
		}
		return nil
	})
}
