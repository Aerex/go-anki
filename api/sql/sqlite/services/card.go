package services

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/aerex/go-anki/api/sql/sqlite/queries"
	repos "github.com/aerex/go-anki/api/sql/sqlite/repositories"
	"github.com/aerex/go-anki/internal/utils"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/google/uuid"
)

type CardService struct {
	cardRepo repos.CardRepo
	colRepo  repos.ColRepo
	deckRepo repos.DeckRepo
	noteRepo repos.NoteRepo
}

func NewCardService(card repos.CardRepo, col repos.ColRepo, deck repos.DeckRepo, note repos.NoteRepo) CardService {
	return CardService{
		cardRepo: card,
		colRepo:  col,
		deckRepo: deck,
		noteRepo: note,
	}
}

// Find will search for a list of cards providing a given Anki query string.
// See https://docs.ankiweb.net/searching.html for more information on querying
func (c *CardService) Find(qs string) (cards []models.Card, err error) {
	var cls string
	var args []string
	if qs != "" {
		bld := queries.NewBuilder(qs)
		cls, args, err = bld.Query()
	}
	// get cards and notes
	cards, err = c.cardRepo.List(cls, args)
	if err != nil {
		return
	}
	noteTypes, err := c.colRepo.NoteTypes()
	if err != nil {
		return
	}
	decks, err := c.deckRepo.Decks()
	if err != nil {
		return
	}
	// join note types and decks to cards
	for i := range cards {
		card := &cards[i]
		card.Note.Model = *noteTypes[card.Note.ModelID]
		card.Deck = *decks[card.DeckID]
	}

	return
}

// Create will create a note and attach to new card to the given deck using the provided
// card type (model)
func (c *CardService) Create(card models.Card, note models.Note, noteType models.NoteType, tmpl models.CardTemplate, deckName string) error {
	decks, err := c.deckRepo.Decks()
	var deckId models.ID
	for _, deck := range decks {
		if deck.Name == deckName {
			deckId = deck.ID
			break
		}
	}
	if err != nil {
		return err
	}
	noteId, err := c.fetchNewId()
	if err != nil {
		return err
	}
	cardId, err := c.fetchNewId()
	if err != nil {
		return err
	}

	usn, err := c.colRepo.USN()
	if err != nil {
		return err
	}
	note.ID = models.ID(noteId)
	note.GUID = uuid.New().String()
	note.ModelID = noteType.ID
	note.USN = usn

	// 1. Check scm to see if we need to do a full sync (use assert?)
	note.SortField = utils.StripHTMLMedia(note.Fields[noteType.SortField])
	csumStr := utils.FieldChecksum(note.Fields[0])
	csum, parseErr := strconv.ParseUint(csumStr[0:8], 16, 64)
	if parseErr != nil {
		return parseErr
	}
	note.Checksum = csum
	card.ID = models.ID(cardId)
	card.NoteID = note.ID

	if card.DeckID == 0 {
		if tmpl.DeckOverride != 0 && decks[tmpl.DeckOverride] != nil {
			card.DeckID = tmpl.DeckOverride
		} else if deckId != 0 {
			card.DeckID = deckId
		}
	}

	deck, exists := decks[card.DeckID]
	if !exists {
		return fmt.Errorf("could not find deck for new card")
	}

	if deck.Dyn {
		card.DeckID = 1
	}
	due, err := c.colRepo.NextDue()
	dconf, err := c.deckRepo.Conf(deck.ID)
	if err != nil {
		return err
	}
	if dconf.New.Order == models.NewCardsDue {
		card.Due = models.UnixTime(due)
	} else {
		// PERF: Some precision lost converting between int64 - float64
		rand.Seed(int64(due))
		due = int64(math.Max(float64(due), 1000))
		card.Due = models.UnixTime(due)
	}

	now := time.Now().Unix()
	card.Mod = models.UnixTime(now)

	card.USN = usn

	// TODO: Insert records into note as well
	// Abstract create note in another method
	// see https://github.com/dae/anki/blob/f1734a475db6f821663b3cf187388d05c3bcc846/pylib/anki/notes.py#L85
	// 2. Check if note exists
	fields := utils.JoinFields(note.Fields)
	err, noteExists := c.noteRepo.Exists(note.ID, note.StringTags, fields)
	if note.Mod == 0 && noteExists {
		return fmt.Errorf("Note %d already exists with tags %s and fields %s",
			note.ID, note.StringTags, strings.Join(note.Fields, ","))
	}
	if note.Mod == 0 {
		note.Mod = card.Mod
	}
	if createNoteErr := c.noteRepo.Create(note); createNoteErr != nil {
		return createNoteErr
	}

	if createCardErr := c.cardRepo.Create(card); createCardErr != nil {
		return createCardErr
	}
	return nil
}

func (c *CardService) fetchNewId() (models.ID, error) {
	id := time.Now()
	// continue to check if new card id exists
	// if so add 1sec and try again
	for {
		err, exists := c.cardRepo.Exists(id.Unix())
		if err != nil {
			return -1, err
		}
		if exists {
			id.Add(time.Second)
		} else {
			break
		}
	}
	return models.ID(id.Unix()), nil
}
