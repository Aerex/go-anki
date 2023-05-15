package queries

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	fanki "github.com/flimzy/anki"

	"github.com/aerex/go-anki/api/sql/sqlite/repositories"
	"github.com/aerex/go-anki/internal/utils"
	"github.com/aerex/go-anki/pkg/models"
)

type mapOrd struct {
	nt models.NoteType
	o  int
}

var NOTE_ID_REGEX = regexp.MustCompile("[^0-9,]")
var READY_CARD_STATE_REGEX = regexp.MustCompile("review|new|learn")
var MODEL_ID_REGEX = regexp.MustCompile("[^0-9]")
var PROP_REGEX = regexp.MustCompile("(^.+?)(<=|>=|!=|=|<|>)(.+?$)")
var VALID_PROPS = []string{"due", "ivl", "reps", "lapses", "ease"}
var FIELD_SEP = "^_"

type clause struct {
	val      string
	cmd      string
	args     []string
	colRepo  repositories.ColRepo
	deckRepo repositories.DeckRepo
	noteRepo repositories.NoteRepo
}

func (c *clause) added() string {
	days, err := strconv.ParseInt(c.val, 10, 0)
	if err != nil {
		return ""
	}
	dayCutoff := c.colRepo.DayCutoff()
	cutoff := (dayCutoff - 86400*days) * 1000
	return fmt.Sprintf("c.id > %d", cutoff)
}

func (c *clause) template() string {
	order, err := strconv.ParseInt(c.val, 10, 0)
	if err != nil {
		order = -1 // use -1 as invalid order
	}
	if order != -1 {
		return fmt.Sprintf("c.ord = %d", order)
	}
	// search for template names
	noteTypes, err := c.colRepo.NoteTypes()
	if err != nil {
		return ""
	}
	var limits []string
	for _, m := range noteTypes {
		for _, tmpl := range m.Templates {
			normTmpl, err := utils.NormalizeString(strings.ToLower(tmpl.Name))
			if err != nil {
				return ""
			}
			if normTmpl == strings.ToLower(c.val) {
				if m.Type == models.ClozeCardType {
					// apply limit if model is cloze
					limits = append(limits, fmt.Sprintf("(n.mid = %d)", m.ID))
				} else {
					limits = append(limits, fmt.Sprintf("(n.mid = %d and c.ord = %d)", m.ID, tmpl.Ordinal))
				}
			}
		}
	}
	return strings.Join(limits, " or ")
}

func (c *clause) deck() string {
	// TODO: Look into migrating into a package
	findDeckByName := func(decks models.Decks, name string) (deck models.Deck, err error) {

		for _, d := range decks {
			if d.Name == name {
				return *d, nil
			}
		}
		return models.Deck{}, err
	}
	decks, err := c.deckRepo.Decks()
	if err != nil {
		return ""
	}
	// skip if searching all decks '*'
	if c.val == "*" {
		return "skip"
	} else if c.val == "filtered" {
		return "c.odid"
	}
	// current deck
	var ids []models.ID
	if strings.ToLower(c.val) == "current" {
		conf, err := c.colRepo.Conf()
		if err != nil {
			return ""
		}
		ids, err = c.deckRepo.DeckIDs(conf.CurrentDeck)
		if err != nil {
			return ""
		}
	} else if !strings.Contains(c.val, "*") {
		d, err := findDeckByName(decks, c.val)
		if err != nil {
			return ""
		}
		ids, err = c.deckRepo.DeckIDs(d.ID)
		if err != nil {
			return ""
		}
	} else {
		// wildcard
		d, err := findDeckByName(decks, c.val)
		if err != nil {
			return ""
		}
		ids, err = c.deckRepo.DeckIDs(d.ID)
		// create a map to ref deck IDs
		var idsMap map[models.ID]bool
		for _, id := range ids {
			idsMap[id] = true
		}
		if err != nil {
			ids = []models.ID{}
			c.val = strings.ReplaceAll(c.val, "*", ".*")
			c.val = strings.ReplaceAll(c.val, "+", "\\+")
			for _, d := range decks {
				deckName, err := utils.NormalizeString(d.Name)
				if err != nil {
					return ""
				}
				reg := regexp.MustCompile("(?i)" + c.val)
				if reg.MatchString(deckName) {
					matchingIds, err := c.deckRepo.DeckIDs(d.ID)
					if err != nil {
						return ""
					}
					for _, id := range matchingIds {
						if _, exists := idsMap[id]; !exists {
							ids = append(ids, id)
						}
					}

					if len(ids) == 0 {
						return ""
					}

					var sids []string
					for _, id := range ids {
						sids = append(sids, fmt.Sprint(id))
					}

					sidls := strings.Join(sids, ", ")
					// convert to sql clauses
					return fmt.Sprintf("c.did in (%s) or c.odid in (%s)", sidls, sidls)
				}
			}
		}
	}

	return ""
}

func (c *clause) flag() string {
	var flag int
	switch c.val {
	case "0":
		flag = 0
		break
	case "1":
		flag = 1
		break
	case "2":
		flag = 2
		break
	case "3":
		flag = 3
		break
	case "5":
		flag = 5
		break
	case "6":
		flag = 6
		break
	case "7":
		flag = 7
		break
	default:
		return ""
	}
	mask := 0b111 //2**3 -1 in Anki
	return fmt.Sprintf("(c.flags & %d) == %d", mask, flag)
}

func (c *clause) mid() string {
	if MODEL_ID_REGEX.MatchString(c.val) {
		return ""
	}
	return fmt.Sprintf("n.mid = %s", c.val)
}

func (c *clause) nid() string {
	if NOTE_ID_REGEX.MatchString(c.val) {
		return ""
	}
	return fmt.Sprintf("n.id in (%s)", c.val)
}

func (c *clause) cid() string {
	if NOTE_ID_REGEX.MatchString(c.val) {
		return ""
	}
	return fmt.Sprintf("c.id in (%s)", c.val)
}

func (c *clause) note() string {
	var sids []string
	models, err := c.colRepo.NoteTypes()
	if err != nil {
		return ""
	}
	for _, model := range models {
		name, err := utils.NormalizeString(strings.ToLower(model.Name))
		if err != nil {
			return ""
		}
		if name == c.val {
			sids = append(sids, string(model.Name))
		}
	}

	if len(sids) == 0 {
		return ""
	}

	return fmt.Sprintf("n.mid in (%s)", strings.Join(sids, ", "))
}

func (c *clause) prop() string {
	if !PROP_REGEX.MatchString(c.val) {
		return ""
	}
	groups := PROP_REGEX.FindStringSubmatch(c.val)
	if len(groups) < 3 {
		// TODO: Need to log if there is an error
		return ""
	}
	prop := strings.ToLower(groups[0])
	cmp := groups[1]
	sval := groups[2]

	var val int64
	if prop == "ease" {
		sf, err := strconv.ParseFloat(sval, 64)
		if err != nil {
			return ""
		}
		val = int64(sf * 10000)
	} else {
		si, err := strconv.ParseInt(sval, 10, 64)
		if err != nil {
			return ""
		}
		val = si
	}

	// validate prop
	if !utils.ArrayStringContains(prop, VALID_PROPS) {
		return ""
	}

	var q []string
	if prop == "due" {
		val += c.colRepo.SchedToday()
		// only valid for review/daily learning
		q = append(q, "(c.queue in (2,3))")
	} else if prop == "ease" {
		prop = "factor"
	}
	q = append(q, fmt.Sprintf("(%s %s %d)", prop, cmp, val))
	return strings.Join(q, " and ")
}

func (c *clause) rated() string {
	rates := strings.Split(c.val, ":")
	days, err := strconv.ParseInt(rates[0], 10, 64)
	if err != nil {
		return ""
	}
	days = int64(math.Min(float64(days), 31))
	var ease string
	if len(rates) > 1 {
		if !utils.ArrayStringContains(rates[1], []string{"1", "2", "3", "4"}) {
			return ""
		}
		ease = fmt.Sprintf("and ease=%s", rates[1])
	}
	cutoff := (c.colRepo.DayCutoff() - 86400*days) * 1000
	return fmt.Sprintf("c.id in (select cid from revlog where id > %d %s)", cutoff, ease)
}

func (c *clause) tag() string {
	val := c.val
	if c.val == "none" {
		return "n.tags = \"\""
	}
	val = strings.ReplaceAll(val, "*", "%")
	if !strings.HasPrefix(val, "%") {
		val = "% " + val
	}
	if !strings.HasSuffix(val, "%") || strings.HasSuffix(val, "\\%") {
		val += " %"
	}
	return ""
}

func (c *clause) dupes() string {
	parts := strings.Split(c.val, ",")
	noteIds := []string{}
	if len(parts) != 2 {
		return ""
	}
	mid, val := parts[0], parts[1]
	csum := utils.FieldChecksum(val)
	notes, err := c.noteRepo.FindByChecksum(mid, csum)
	if err != nil {
		return ""
	}
	for notes.Next() {
		var note models.Note
		notes.Scan(&note)
		if utils.StripHTMLMedia(note.Fields[0]) == val {
			noteIds = append(noteIds, string(note.ID))
		}
	}
	return fmt.Sprintf("n.id in (%s)", strings.Join(noteIds, ","))
}

func (c *clause) cardState() string {
	var cardType models.CardType
	if READY_CARD_STATE_REGEX.MatchString(c.val) {
		switch c.val {
		case "review":
			cardType = models.CardTypeReview
			break
		case "new":
			cardType = models.CardTypeNew
		default:
			return fmt.Sprintf("queue in (1, %d)", models.CardQueueRelearning)
		}
		return "type = " + string(cardType)
	} else if c.val == "suspended" {
		return fmt.Sprintf("c.queue = %d", models.CardQueueSuspended)
	} else if c.val == "buried" {
		return fmt.Sprintf("c.queue in (%d, %d)", models.CardQueueBuried, models.CardQueueSBuried)
	} else if c.val == "due" {
		return fmt.Sprintf("(c.queue in (%d, %d) and c.due <= %d", fanki.CardQueueReview, fanki.CardQueueRelearning, c.colRepo.SchedToday())
	}
	return ""
}

func (c *clause) field() string {
	val := strings.ReplaceAll(c.val, "*", "%")
	noteTypes, err := c.colRepo.NoteTypes()
	if err != nil {
		return ""
	}

	var modelIds []string
	var noteIds []string
	modelToOrderMap := make(map[models.ID]mapOrd)
	for _, noteType := range noteTypes {
		for _, fld := range noteType.Fields {
			fieldName, err := utils.NormalizeString(fld.Name)
			if err != nil {
				return ""
			}
			if strings.ToLower(c.cmd) == fieldName {
				if _, exists := modelToOrderMap[noteType.ID]; !exists {
					modelIds = append(modelIds, string(noteType.ID))
				}
				modelToOrderMap[noteType.ID] = mapOrd{nt: *noteType, o: fld.Ordinal}
			}
		}
	}

	if len(modelToOrderMap) == 0 {
		// nothing has that field
		return ""
	}

	jsVal := strings.ReplaceAll(regexp.QuoteMeta(val), "_", ".")
	jsVal = strings.ReplaceAll(jsVal, regexp.QuoteMeta("%"), ".*")
	jsValReg, err := regexp.Compile(fmt.Sprintf("(?si)^%s%", jsVal))
	if err != nil {
		return ""
	}

	notes, err := c.noteRepo.FindByModelIdsField(val, modelIds)
	for notes.Next() {
		var note models.Note
		notes.Scan(&note)
		modelToOrd, exists := modelToOrderMap[note.ModelID]
		if !exists {
			return ""
		}
		if jsValReg.MatchString(note.Fields[modelToOrd.o]) {
			noteIds = append(noteIds, string(note.ID))
		}
	}
	if len(noteIds) == 0 {
		return "0"
	}
	return fmt.Sprintf("n.id in (%s)", strings.Join(noteIds, ","))
}

func (c *clause) text(token string) string {
	val := strings.ReplaceAll(token, "*", "%")
	c.args = append(c.args, "%"+val+"%")
	c.args = append(c.args, "%"+val+"%")
	return "note.sfld like ? escape \"\\\" or note.flds like ? escape \"\\\""
}
