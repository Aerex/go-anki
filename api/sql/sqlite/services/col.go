package services

import (
	"sort"

	repos "github.com/aerex/go-anki/api/sql/sqlite/repositories"
	"github.com/aerex/go-anki/pkg/models"
)

type ColService struct {
	colRepo repos.ColRepo
}

func NewColService(c repos.ColRepo) ColService {
	return ColService{
		colRepo: c,
	}
}

// NoteTypeByName returns the NoteType given the name
func (c *ColService) GetNoteTypeByName(name string) (noteType models.NoteType, err error) {
	noteTypes, err := c.colRepo.NoteTypes()
	if err != nil {
		return
	}
	for _, ntype := range noteTypes {
		if ntype.Name == name {
			return *ntype, err
		}
	}
	return
}

func (c *ColService) NoteTypes() (models.NoteTypes, error) {
	noteTypes, err := c.colRepo.NoteTypes()
	if err != nil {
		return models.NoteTypes{}, err
	}
	// always have card fields sorted by ordinal property
	for _, nt := range noteTypes {
		sort.Sort(repos.ByOrdinal(nt.Fields))
	}
	return noteTypes, nil
}

func (c *ColService) Tags() ([]string, error) {
	return c.colRepo.Tags()
}
