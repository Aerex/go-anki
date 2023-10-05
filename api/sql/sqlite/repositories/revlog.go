package repositories

import (
	"github.com/aerex/go-anki/pkg/models"
	"github.com/jmoiron/sqlx"
)

type RevLogRepo interface {
	TodayStats(dayCutoff int64) (stats models.StudiedToday, err error)
	MaturedCards(dayCutoff int64) (stats models.MaturedToday, err error)
}

type revLogRepo struct {
	Conn *sqlx.DB
	Tx   *sqlx.Tx
}

func NewRevLogRepository(conn *sqlx.DB) RevLogRepo {
	return revLogRepo{
		Conn: conn,
	}
}

func (r revLogRepo) TodayStats(dayCutoff int64) (stats models.StudiedToday, err error) {
	query := `SELECT COUNT() "cards", SUM(time)/1000 "time",
    SUM(CASE WHEN ease = 1 THEN 1 ELSE 0 END) "failed",
    SUM(CASE WHEN type = 0 THEN 1 ELSE 0 END) "learning",
    SUM(CASE WHEN type = 1 THEN 1 ELSE 0 END) "review",
    SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) "relearned,
    SUM(CASE WHEN type = 3 THEN 1 ELSE 0 END) "filter"
      FROM revlog WHERE id > ?`

	if err = r.Conn.Get(&stats, query, cutoff(dayCutoff)); err != nil {
		return
	}
	return
}

func (r revLogRepo) MaturedCards(dayCutoff int64) (stats models.MaturedToday, err error) {
	query := `SELECT COUNT() "mcount",
              SUM(CASE WHEN ease = 1 THEN 0 ELSE 1 END) FROM revlog
                WHERE lastIvl >= 21 AND id > ?`

	if err = r.Conn.Get(&stats, query, cutoff(dayCutoff)); err != nil {
		return
	}
	return
}

func cutoff(dayCutoff int64) int64 {
	return (dayCutoff - 86400) * 1000
}
