package repositories

import (
	"fmt"
	"time"

	ankisql "github.com/aerex/go-anki/api/sql"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/maps"
)

type ByOrdinal []*models.CardField

func (n ByOrdinal) Len() int           { return len(n) }
func (n ByOrdinal) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n ByOrdinal) Less(i, j int) bool { return n[i].Ordinal < n[j].Ordinal }

type colRepo struct {
	Conn *sqlx.DB
	Tx   ankisql.TxOpts
}

type ColRepo interface {
	Conf() (conf models.CollectionConf, err error)
	UpdateMod() (err error)
	CreatedTime() (crt models.UnixTime, err error)
	NextDue() (due int64, err error)
	DeckConf(deckId models.ID) (deckConf models.DeckConfig, err error)
	SchedToday() int64
	USN(server bool) (usn int, err error)
	Rollover() int
	DayCutoff() int64
	Tags() (tags []string, err error)
	NoteTypes() (noteTypes models.NoteTypes, err error)
}

func NewColRepository(conn *sqlx.DB) ColRepo {
	return colRepo{
		Conn: conn,
		Tx: ankisql.TxOpts{
			DB: conn,
		},
	}
}
func (c colRepo) UpdateMod() (err error) {
	return ankisql.Tx(c.Tx, func(tx *sqlx.Tx) error {
		query := fmt.Sprintf("UPDATE col SET mod = %d WHERE ID = 1", time.Now().Unix())
		if _, err := tx.Exec(query); err != nil {
			return err
		}
		return nil
	})
}

func (c colRepo) CreatedTime() (crt models.UnixTime, err error) {
	query := `SELECT crt FROM col`
	if err = c.Conn.Get(&crt, query); err != nil {
		return
	}
	return
}

func (c colRepo) Conf() (conf models.CollectionConf, err error) {
	var col models.Collection
	query := `SELECT conf FROM col`
	if err = c.Conn.Get(&col, query); err != nil {
		return
	}
	conf = col.Conf
	return
}

// NextDue will retrieve the next expected due value for a card
// The due value is calculated by adding 1 to the nextPos property defined in col.conf
func (c colRepo) NextDue() (due int64, err error) {
	var colConf models.CollectionConf
	colConf, err = c.Conf()
	if err != nil {
		return
	}
	return int64(colConf.NextPos + 1), nil
}

func (c colRepo) DeckConf(deckId models.ID) (deckConf models.DeckConfig, err error) {
	var deckConfs models.DeckConfigs
	query := "SELECT dconf FROM col LIMIT 1"
	if err = c.Conn.QueryRowx(query).Scan(&deckConfs); err != nil {
		fmt.Printf("query: %s", err.Error())
		return
	}
	deckConf = *deckConfs[deckId]

	return
}

// USN retrieves the update sequence of col
func (c colRepo) USN(server bool) (usn int, err error) {
	if server {
		var col models.Collection
		query := `SELECT usn FROM col`
		if err = c.Conn.Get(&col, query); err != nil {
			return
		}
		usn = col.USN
		return
	}
	return -1, nil
}

func (c colRepo) Rollover() int {
	// TODO: Figure how to get rolloverTime; Default to 4 for now
	// rollover is not found in db or anywhere that I can see
	// c.self.col.conf.get("rollover", 4)
	return 4
}

func (c colRepo) SchedToday() int64 {
	query := `SELECT crt FROM col`
	var crt int64
	if err := c.Conn.Select(crt, query); err != nil {
		return time.Now().Unix()
	}
	return time.Now().Unix() - crt
}

func (c colRepo) DayCutoff() int64 {
	rolloverTime := c.Rollover()
	if rolloverTime < 0 {
		rolloverTime = 24 + rolloverTime
	}
	date := time.Now()
	date = time.Date(date.Year(), date.Month(), 0, rolloverTime, 0, 0, 0, date.Location())
	if date.Before(time.Now()) {
		date = date.Add(time.Hour * 24)
	}
	return date.Unix()
}

func (c colRepo) NoteTypes() (noteTypes models.NoteTypes, err error) {
	query := `SELECT models FROM col LIMIT 1`
	if err = c.Conn.QueryRowx(query).Scan(&noteTypes); err != nil {
		return
	}
	return
}

func (c colRepo) Tags() (tags []string, err error) {
	var tagCache models.TagCache
	query := `SELECT tags From col LIMIT 1`
	if err = c.Conn.QueryRowx(query).Scan(&tagCache); err != nil {
		return
	}
	tags = maps.Keys(tagCache)

	return

}
