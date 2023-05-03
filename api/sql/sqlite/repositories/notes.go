package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/aerex/go-anki/pkg/models"
	fanki "github.com/flimzy/anki"

	"github.com/jmoiron/sqlx"
)

type noteRepo struct {
	Conn *sqlx.DB
	Tx   *sqlx.Tx
}

type NoteRepo interface {
	FindByChecksum(mid, csum string) (notes fanki.Notes, err error)
	FindByModelIdsField(field string, mids []string) (notes fanki.Notes, err error)
	FindById(id string) (note fanki.Note, err error)
	Create(note models.Note) (err error)
	Exists(id models.ID, stringTags string, fields string) (err error, exists bool)
	WithTrans(trans interface{}) NoteRepo
	MustCreateTrans() *sqlx.Tx
}

func NewNoteRepository(conn *sqlx.DB) NoteRepo {
	return noteRepo{
		Conn: conn,
	}
}

func (n noteRepo) FindById(id string) (note fanki.Note, err error) {
	query := `SELECT * FROM notes WHERE id=? LIMIT 1`
	if err = n.Conn.Get(&note, query, id); err != nil {
		return
	}
	return
}

func (n noteRepo) FindByChecksum(mid, csum string) (notes fanki.Notes, err error) {
	query := `SELECT id, flds FROM notes WHERE mid=? and csum=?`
	if err = n.Conn.Select(notes, query, mid, csum); err != nil {
		return
	}
	return
}

func (n noteRepo) FindByModelIdsField(field string, mids []string) (notes fanki.Notes, err error) {
	query := fmt.Sprintf("SELECT id, mid, flds FROM notes WHERE mid in (%s) and flds like ? escape '\\'", strings.Join(mids, ","))
	query = query + " and flds like ? escape '\\'"
	if err = n.Conn.Select(notes, query, "%"+field+"%"); err != nil {
		return
	}
	return
}

func (n noteRepo) Exists(id models.ID, stringTags string, fields string) (err error, exists bool) {
	var note models.Note
	query := "SELECT 1 from notes WHERE ID = ? AND tags = ? AND flds = ?"
	err = n.Conn.Get(&note, query, id, stringTags, fields)
	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		return err, false
	}
	return
}

func (n noteRepo) Create(note models.Note) (err error) {
	query := `INSERT OR REPLACE INTO notes (id, guid, mid, mod, usn, tags, flds, sfld, csum, flags, data)
    VALUES (?,?,?,?,?,?,?,?,?,?, "")`
	if n.Tx != nil {
		if _, err = n.Tx.Exec(query, note.ID, note.GUID, note.ModelID, note.Mod,
			note.USN, note.StringTags, note.Fields, note.SortField, note.Checksum, note.Flags); err != nil {
			return
		}
		return
	}
	if _, err = n.Conn.NamedExec(query, note); err != nil {
		return
	}
	return
}

func (n noteRepo) WithTrans(trans interface{}) NoteRepo {
	n.Tx = trans.(*sqlx.Tx)
	return n
}

func (n noteRepo) MustCreateTrans() *sqlx.Tx {
	return n.Conn.MustBegin()
}
