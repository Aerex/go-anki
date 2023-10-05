package sql

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/aerex/go-anki/pkg/models"
)

type TxOpts struct {
	DB     *sqlx.DB
	DryRun bool
}

func InClauseFromIDs(IDs []models.ID) string {
	var ids []string
	for _, id := range IDs {
		ids = append(ids, fmt.Sprint(id))
	}

	return "(" + strings.Join(ids, ",") + ")"

}

func Tx(opts TxOpts, cb func(tx *sqlx.Tx) error) error {
	tx := opts.DB.MustBegin()
	if err := cb(tx); err != nil {
		return err
	}

	if !opts.DryRun {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
