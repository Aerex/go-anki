package v2

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"golang.org/x/exp/maps"

	"github.com/google/gapid/core/math/sint"
	"github.com/op/go-logging"

	"github.com/aerex/go-anki/api/sql"
	repos "github.com/aerex/go-anki/api/sql/sqlite/repositories"
	"github.com/aerex/go-anki/internal/utils"
	"github.com/aerex/go-anki/pkg/models"
)

type ByLearnDue []*learnQueue

func (d ByLearnDue) Len() int           { return len(d) }
func (d ByLearnDue) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d ByLearnDue) Less(i, j int) bool { return d[i].Due < d[j].Due }

type learnQueue struct {
	Due models.UnixTime
	ID  models.ID
}

type schedV2Service struct {
	colRepo           repos.ColRepo
	deckRepo          repos.DeckRepo
	cardsRepo         repos.CardRepo
	revLogRepo        repos.RevLogRepo
	noteRepo          repos.NoteRepo
	colConf           *models.CollectionConf
	server            bool
	revCount          int
	revQueue          []models.ID
	newCount          int
	newCardModulus    int
	dayCutoff         int64
	lrnCutoff         int64
	lrnQueue          []*learnQueue
	lrnDayQueue       []int
	lrnDeckIDs        []models.ID
	newDeckIDs        []models.ID
	newQueue          []models.ID
	haveQueues        bool
	burySiblingsOnAns bool
	learningCount     int
	today             int64
}

var (
	logger               = logging.MustGetLogger("ankicli")
	FactorAdditionValues = []int64{-150, 0, 150}
)

const (
	DynReportLimit = 99999
	ReportLimit    = 1000
)

type SchedService interface {
	DeckStudyStats() (map[models.ID]models.DeckStudyStats, error)
	AnswerButtons(card models.Card) (int, error)
	AnswerCard(card models.Card, ease models.Ease) error
	NextIntervalString(card models.Card, ease models.Ease, conf models.DeckConfig) (string, error)
	//BuryCards(cardIDs []models.Card)
}

func NewSchedV2Service(c repos.ColRepo, cd repos.CardRepo, d repos.DeckRepo, r repos.RevLogRepo, n repos.NoteRepo, server bool) SchedService {
	return schedV2Service{
		colRepo:           c,
		revLogRepo:        r,
		deckRepo:          d,
		cardsRepo:         cd,
		noteRepo:          n,
		server:            server,
		burySiblingsOnAns: true,
	}
}

func (s schedV2Service) DeckStudyStats() (map[models.ID]models.DeckStudyStats, error) {
	stats := make(map[models.ID]models.DeckStudyStats)
	if err := s.checkDay(); err != nil {
		return stats, err
	}
	deckMap, err := s.deckRepo.Decks()
	if err != nil {
		return stats, err
	}
	deckIDs := maps.Keys(deckMap)
	decks := maps.Values(deckMap)
	sort.Sort(repos.ByDeckName(decks))
	deckIDsInClause := sql.InClauseFromIDs(deckIDs)
	if err = s.cardsRepo.RecoverOrphans(deckIDsInClause); err != nil {
		return stats, err
	}
	if err = s.colRepo.UpdateMod(); err != nil {
		return stats, err
	}
	usn, err := s.colRepo.USN(s.server)
	if err != nil {
		return stats, err
	}
	err = s.deckRepo.FixDecks(deckMap, usn)
	if err != nil {
		return stats, err
	}
	limits := make(map[string][]int)

	var nlmt int
	var plmt int
	for _, deck := range decks {
		p := parent(deck.Name)
		// new
		nlmt, err = s.deckLimitForNewCards(*deck)
		if err != nil {
			return stats, err
		}
		if p != "" {
			nlmt = sint.Min(nlmt, limits[p][0])
		}
		newCardCount, err := s.cardsRepo.CardsNewForDeck(deck.ID, nlmt)
		if err != nil {
			return stats, err
		}
		colConf, err := s.colRepo.Conf()
		if err != nil {
			return stats, err
		}
		due := time.Now().Unix() + int64(colConf.CollapseTime)
		lrnCardCnt, err := s.cardsRepo.CardsLearnedForDeck(int64(deck.ID), due, s.today, ReportLimit)
		if err != nil {
			return stats, err
		}
		if p != "" {
			plmt = limits[p][1]
		} else {
			plmt = -1
		}
		reviewLmts, err := s.deckLimitForReviewCards(*deck, plmt)
		if err != nil {
			return stats, err
		}
		childIDs, err := s.deckRepo.ChildrenDeckIDs(deck.ID)
		if err != nil {
			return stats, err
		}

		dids := append(childIDs, deck.ID)
		reviewCardCnt, err := s.cardsRepo.CardsReviewForDeck(sql.InClauseFromIDs(dids), ReportLimit, reviewLmts, int(s.today))
		if err != nil {
			return stats, err
		}
		stats[deck.ID] = models.DeckStudyStats{
			New:      newCardCount,
			Review:   reviewCardCnt,
			Learning: lrnCardCnt,
		}
	}
	return stats, err
}

// CheckDay will check if the day has rolled over
// passed the cutoff day. If so, reset
func (s *schedV2Service) checkDay() error {
	cutoff := s.colRepo.DayCutoff()
	if time.Now().Unix() > cutoff {
		if err := s.reset(false); err != nil {
			return err
		}
	}
	return nil
}

func (s *schedV2Service) reset(server bool) error {
	if err := s.updateCutoff(); err != nil {
		return err
	}
	if err := s.resetLrn(); err != nil {
		return err
	}
	if err := s.resetRev(); err != nil {
		return err
	}
	if err := s.resetNew(); err != nil {
		return err
	}
	s.haveQueues = true
	return nil
}

func (s *schedV2Service) currentTimezoneOffset() (int32, error) {
	if s.server {
		conf, err := s.colRepo.Conf()
		if err != nil {
			return 0, err
		}
		return conf.LocalOffset, nil
	} else {
		now := time.Now()
		_, offset := now.Zone()
		return int32((offset * -1) / 60), nil
	}
}
func fixedOffsetFromMin(minWest int32) *time.Location {
	boundedCrtMin := sint.Max(-23*60, int(minWest))
	boundedCrtMin = sint.Min(23*60, boundedCrtMin)
	// TODO: Confirm if UTC-12 is western hemisphere
	return time.FixedZone("UTC-12", (boundedCrtMin * 60))
}

func normalizedRollowedHour(hour int) int {
	cappedHour := sint.Max(hour, -23)
	cappedHour = sint.Min(cappedHour, 23)
	if cappedHour < 0 {
		return 24 + cappedHour
	}
	return cappedHour
}

func daysElapsed(startDate time.Time, endDate time.Time, rolloverPassed bool) int64 {
	days := (endDate.Sub(startDate).Abs().Hours()) / 24

	if rolloverPassed {
		return int64(days)
	}
	return int64(days - 1)
}

func (s *schedV2Service) timingToday(crt models.UnixTime, crtMinWest int32, nowMinWest int32, rolloverHr int) models.SchedTimingToday {
	createdDate := time.Unix(int64(crt), 0).In(fixedOffsetFromMin(crtMinWest))
	currentDate := time.Now().In(fixedOffsetFromMin(nowMinWest))

	rolloverHr = normalizedRollowedHour(rolloverHr)
	rolloverTodayDateTime := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), rolloverHr, currentDate.Minute(), currentDate.Second(), currentDate.Nanosecond(), fixedOffsetFromMin(nowMinWest))
	nextDateAt := rolloverTodayDateTime.Unix()
	rolloverPassed := rolloverTodayDateTime.Before(currentDate)
	if rolloverPassed {
		nextDateAt = rolloverTodayDateTime.Add(24 * time.Hour).Unix()
	}

	daysElapsed := daysElapsed(createdDate, currentDate, rolloverPassed)
	return models.SchedTimingToday{
		DaysElapsed: daysElapsed,
		NextDayAt:   nextDateAt,
	}
}
func (s *schedV2Service) _dayCutoff(colConf models.CollectionConf) int64 {
	rollover := colConf.Rollover
	if rollover == 0 {
		rollover = 4
	}
	if rollover < 0 {
		rollover = 24 + rollover
	}

	date := time.Now()
	date = time.Date(date.Year(), date.Month(), date.Day(), int(rollover), 0, 0, 0, time.UTC)
	if date.Before(time.Now()) {
		date = date.AddDate(0, 0, 1)
	}
	return date.Unix()
}

func (s *schedV2Service) daysSinceCreation(crt models.UnixTime, rollover int) int64 {
	start := time.Unix(int64(crt), 0)
	start = time.Date(start.Year(), start.Month(), start.Day(), rollover, 0, 0, 0, time.UTC)
	return int64((time.Now().Unix() - start.Unix()) / 86400)
}

func (s *schedV2Service) updateCutoff() error {
	colConf, err := s.colRepo.Conf()
	if err != nil {
		return err
	}
	oldToday := s.today
	createdTime, err := s.colRepo.CreatedTime()
	if err != nil {
		return err
	}
	offset, err := s.currentTimezoneOffset()
	if err != nil {
		return err
	}
	timing := s.timingToday(createdTime, colConf.CreationOffset, offset, int(colConf.Rollover))
	if colConf.CreationOffset != 0 {
		s.today = timing.DaysElapsed
		s.dayCutoff = timing.NextDayAt
	} else {
		s.today = s.daysSinceCreation(createdTime, int(colConf.Rollover))
		s.dayCutoff = s._dayCutoff(colConf)
	}
	if oldToday != s.today {
		// self.col.log(self.today, self.dayCutoff)
	}

	decks, err := s.deckRepo.Decks()
	if err != nil {
		return err
	}
	for _, deck := range decks {
		updateDeck(deck, s.today)
	}
	// unbury if the day has rolled over
	if colConf.LastUnburied < s.today {
		if err := s.cardsRepo.UnburyCards(); err != nil {
			return err
		}
	}
	return nil
}

func (s *schedV2Service) resetRev() error {
	if err := s.resetRevCount(); err != nil {
		return err
	}
	s.revQueue = []models.ID{}
	return nil
}

func (s *schedV2Service) resetRevCount() error {
	colConf, err := s.colRepo.Conf()
	if err != nil {
		return err
	}
	limit, err := s.currentRevLimit()
	if err != nil {
		return err
	}

	deckLimit := sql.InClauseFromIDs(colConf.ActiveDecks)
	revisions, err := s.cardsRepo.Revisions(deckLimit, limit)
	if err != nil {
		return err
	}
	s.revCount = revisions

	return nil
}

func (s *schedV2Service) currentRevLimit() (int, error) {
	colConf, err := s.colRepo.Conf()
	if err != nil {
		return 0, err
	}
	decks, err := s.deckRepo.Decks()
	if err != nil {
		return 0, err
	}
	selectedDeck, exists := decks[models.ID(colConf.CurrentDeck)]
	if !exists {
		// TODO: log deck does not exist
		return 0, fmt.Errorf("deck %d could not be found", colConf.CurrentDeck)
	}
	return s.deckRevLimit(*selectedDeck, -1), nil
}

func (s *schedV2Service) deckRevLimit(deck models.Deck, parentLimit int) int {
	if deck.ID == 0 {
		return 0
	}
	if deck.Dyn {
		return DynReportLimit
	}
	deckConf, err := s.colRepo.DeckConf(models.ID(deck.Conf))
	if err != nil {
		// TODO: log error
		return 0
	}
	limit := sint.Max(0, (deckConf.Rev.PerDay - int(deck.ReviewsToday[1])))

	if parentLimit != -1 {
		return sint.Min(parentLimit, limit)
	} else if !strings.Contains(deck.Name, "::") {
		return limit
	}
	deckParents, err := s.deckRepo.Parents(deck.ID)
	if err != nil {
		// TODO: log error
		return 0
	}
	for _, parent := range deckParents {
		limit = sint.Min(limit, s.deckRevLimit(parent, limit))
	}
	return limit
}

// deckLimitForNewCards get the limit for deck without parent limits
func (s *schedV2Service) deckLimitForNewCards(deck models.Deck) (int, error) {
	if deck.Dyn {
		return DynReportLimit, nil
	}
	conf, err := s.deckRepo.Conf(models.ID(deck.Conf))
	if err != nil {
		return 0, err
	}
	return sint.Max(0, conf.New.PerDay-int(deck.NewToday[1])), nil
}

func (s *schedV2Service) deckLimitForReviewCards(deck models.Deck, parentLimit int) (int, error) {
	if deck.Dyn {
		return DynReportLimit, nil
	}
	conf, err := s.deckRepo.Conf(models.ID(deck.Conf))
	if err != nil {
		return 0, err
	}
	limit := sint.Max(0, conf.Rev.PerDay-int(deck.ReviewsToday[1]))
	if parentLimit != -1 {
		return sint.Min(parentLimit, limit), nil
	} else if !strings.Contains(deck.Name, "::") {
		return limit, nil
	}
	parents, err := s.deckRepo.Parents(deck.ID)
	if err != nil {
		return 0, err
	}
	for _, parent := range parents {
		plim, err := s.deckLimitForReviewCards(parent, limit)
		if err != nil {
			return 0, err
		}
		limit = sint.Min(limit, plim)
		if err != nil {
			return 0, err
		}
	}
	return limit, nil
}

func (s *schedV2Service) resetNew() error {
	colConf, err := s.colRepo.Conf()
	if err != nil {
		return err
	}
	newCount, err := s.resetNewCount()
	if err != nil {
		return err
	}
	s.newCount = newCount
	s.newDeckIDs = make([]models.ID, len(colConf.ActiveDecks))
	copy(s.newDeckIDs, colConf.ActiveDecks)
	s.newQueue = []models.ID{}
	s.updateNewCardRatio(colConf)
	return nil
}

func (s *schedV2Service) resetNewCount() (int, error) {
	return s.computeCount(s.deckLimitForNewCards, s.cardsRepo.CardsNewForDeck)
}

func (s *schedV2Service) computeCount(lmtCb func(deck models.Deck) (int, error), cntCb func(deckID models.ID, limit int) (int, error)) (count int, err error) {
	colConf, err := s.colRepo.Conf()
	if err != nil {
		return 0, err
	}
	decks, err := s.deckRepo.Decks()
	if err != nil {
		return 0, err
	}
	uniqueParentCounts := make(map[models.ID]int)
	for _, id := range colConf.ActiveDecks {
		deckID := models.ID(id)
		deck, exists := decks[deckID]
		if !exists {
			deck = decks[models.ID(1)]
		}
		limit, err := lmtCb(*deck)
		if err != nil {
			return 0, err
		}
		if limit < 1 {
			continue
		}
		parentDecks, err := s.deckRepo.Parents(deckID)
		if err != nil {
			return 0, err
		}

		for _, parentDeck := range parentDecks {
			_, exists := uniqueParentCounts[parentDeck.ID]
			if !exists {
				uniqueParentCounts[parentDeck.ID], err = lmtCb(parentDeck)
				if err != nil {
					return 0, err
				}
			}
			limit = sint.Min(uniqueParentCounts[parentDeck.ID], limit)
		}
		curCount, err := cntCb(deckID, limit)
		if err != nil {
			return 0, err
		}
		for parentDeck := range parentDecks {
			uniqueParentCounts[models.ID(parentDeck)] -= curCount
		}
		uniqueParentCounts[deckID] = limit - curCount
		count += curCount
	}
	return
}

func (s *schedV2Service) resetLrn() error {
	colConf, err := s.colRepo.Conf()
	if err != nil {
		return err
	}
	s.updateLrnCutoff(true, colConf)
	s.resetLrnCount(colConf)
	s.lrnQueue = []*learnQueue{}
	s.lrnDayQueue = []int{}
	s.lrnDeckIDs = colConf.ActiveDecks
	return nil
}

func (s *schedV2Service) updateLrnCutoff(force bool, colConf models.CollectionConf) bool {
	nextCutoff := time.Now().Unix() + int64(colConf.CollapseTime)
	if ((nextCutoff - s.lrnCutoff) > 60) || force {
		s.lrnCutoff = nextCutoff
		return true
	}
	return false
}

func (s *schedV2Service) resetLrnCount(colConf models.CollectionConf) error {
	deckLimit := sql.InClauseFromIDs(colConf.ActiveDecks)

	learningCount, err := s.cardsRepo.LearningCount(deckLimit, s.lrnCutoff)
	if err != nil {
		return err
	}
	s.learningCount = learningCount
	return nil
}

func (s *schedV2Service) updateNewCardRatio(colConf models.CollectionConf) {
	if colConf.NewSpread == models.NewCardSpread(models.NewCardsDistribute) {
		if s.newCount > 0 {
			s.newCardModulus = (s.newCount + s.revCount) / s.newCount
			if s.revCount > 0 {
				s.newCardModulus = sint.Max(2, s.newCardModulus)
			}

		}
		return
	}
	s.newCardModulus = 0
}

// AnswerButtons returns the number of buttons to show when studying a deck
func (s schedV2Service) AnswerButtons(card models.Card) (int, error) {
	deckConf, err := s.deckRepo.Conf(card.DeckID)
	if err != nil {
		return 0, err
	}
	if card.OriginalDeckID != 0 && !deckConf.Resched {
		return 2, nil
	}
	return 4, nil
}

func (s schedV2Service) AnswerCard(card models.Card, ease models.Ease) error {
	deckConf, err := s.deckRepo.Conf(card.DeckID)
	if err != nil {
		return err
	}
	// TODO: figure out how to do markReview
	// @see anki pylib/anki/schedv2.py
	if s.burySiblingsOnAns {
		if err := s.burySiblings(card, deckConf); err != nil {
			return err
		}
	}
	previewCard, err := s.previewingCard(card)
	if err != nil {
		return err
	}
	if previewCard {
		card, err = s.answerCardPreview(card, ease, deckConf)
		if err != nil {
			return err
		}
		if err := s.updateStats(card, models.CardTypeTime, TimeTaken(card, deckConf)); err != nil {
			return err
		}
	} else {
		card.Reps = card.Reps + 1
		// move new cards in queue to learning
		if card.Queue == models.CardQueueNew {
			card.Queue = models.CardQueueLearning
			card.Type = models.CardTypeLearning
			left, err := s.startingLeft(card, deckConf)
			if err != nil {
				return err
			}
			// initital reviews to complete
			card.ReviewsLeft = left
			if err := s.updateStats(card, models.CardTypeNew, 1); err != nil {
				return err
			}
		}
		if card.Queue == models.CardQueueLearning || card.Queue == models.CardQueueRelearning {
			if err := s.answerLearnCard(card, deckConf, ease); err != nil {
				return err
			}
		} else if card.Queue == models.CardQueueReview {
			if err := s.answerReviewCard(card, ease, deckConf); err != nil {
				return err
			}
			if err := s.updateStats(card, models.CardTypeReview, 1); err != nil {
				return err
			}
		}

		if card.OriginalDue != 0 {
			card.OriginalDue = 0
		}
	}
  usn, err := s.colRepo.USN(false)
  if err != nil {
    return err
  }
  card.Mod = models.UnixTime(time.Now().Unix())
  card.USN = usn
  if err := s.cardsRepo.Update(card); err != nil {
    return err
  }
	return nil
}

func (s schedV2Service) emptyDyn(deckID models.ID, lim string) error {
	if lim == "" {
		lim = "did = " + fmt.Sprint(deckID)
	}
	usn, err := s.colRepo.USN(true)
	if err != nil {
		return err
	}
	return s.cardsRepo.EmptyDyn(lim, usn)
}

func (s schedV2Service) burySiblings(card models.Card, deckConfig models.DeckConfig) error {
	var cardsToBury []models.ID
	newCardConf, err := s.newCardConf(card, deckConfig)
	if err != nil {
		return err
	}
	var buryNew bool
	if newCardConf.Bury == nil {
		buryNew = true
	} else {
		buryNew = *newCardConf.Bury
	}
	revConf, err := s.revCardConf(card)
	if err != nil {
		return err
	}
	var buryRev bool
	if revConf.Bury == nil {
		buryRev = true
	} else {
		buryRev = *revConf.Bury
	}
	buriedCards, err := s.cardsRepo.BuriedCards(card.NoteID, card.ID)
	if err != nil {
		return err
	}
	for _, bcard := range buriedCards {
		if bcard.Queue == models.CardQueueReview {
			if buryRev {
				cardsToBury = append(cardsToBury, bcard.ID)
			}
			_, s.revQueue = utils.DequeueModelID(s.revQueue, &bcard.ID)
		} else {
			if buryNew {
				cardsToBury = append(cardsToBury, bcard.ID)
			}
			_, s.newQueue = utils.DequeueModelID(s.newQueue, &bcard.ID)
		}
		if len(cardsToBury) > 0 {
			usn, err := s.colRepo.USN(true)
			if err != nil {
				return err
			}
			s.cardsRepo.BuryCards(cardsToBury, usn)
		}
	}
	return nil
}

func (s schedV2Service) newCardConf(card models.Card, deckConf models.DeckConfig) (models.NewDeckConf, error) {
	// normal deck
	if card.OriginalDeckID == 0 {
		return deckConf.New, nil
	}
	// dynamic deck
	origConf, err := s.deckRepo.Conf(card.OriginalDeckID)
	if err != nil {
		return models.NewDeckConf{}, err
	}
	var bury bool
	if origConf.New.Bury != nil {
		bury = *origConf.New.Bury
	} else {
		bury = true
	}
	return models.NewDeckConf{
		Bury:          &bury,
		Ints:          origConf.New.Ints,
		Delays:        origConf.New.Delays,
		InitialFactor: origConf.New.InitialFactor,
		Order:         models.NewCardsDue,
		Seperate:      origConf.New.Seperate,
		PerDay:        ReportLimit,
	}, nil
}

func (s schedV2Service) revCardConf(card models.Card) (models.RevDeckConf, error) {
	conf, err := s.deckRepo.Conf(card.DeckID)
	if err != nil {
		return models.RevDeckConf{}, err
	}

	// normal deck
	if card.OriginalDeckID == 0 {
		return conf.Rev, nil
	}
	// dynamic deck
	origConf, err := s.deckRepo.Conf(card.OriginalDeckID)
	if err != nil {
		return models.RevDeckConf{}, err
	}
	return origConf.Rev, nil
}

func (s schedV2Service) lapseCardConf(card models.Card, deckConf models.DeckConfig) (models.LapseDeckConf, error) {
	if card.OriginalDeckID == 0 {
		return deckConf.Lapse, nil
	}
	// dynamic deck
	origConf, err := s.deckRepo.Conf(card.OriginalDeckID)
	if err != nil {
		return models.LapseDeckConf{}, err
	}
	return models.LapseDeckConf{
		Delays:      origConf.Lapse.Delays,
		LeechAction: origConf.Lapse.LeechAction,
		LeechFails:  origConf.Lapse.LeechFails,
		MinInterval: origConf.Lapse.MinInterval,
		Mult:        origConf.Lapse.MinInterval,
		Resched:     origConf.Lapse.Resched,
	}, nil
}

func (s schedV2Service) learnCardConf(card models.Card, deckConf models.DeckConfig) (interface{}, error) {
	if card.Type == models.CardTypeReview || card.Type == models.CardTypeRelearning {
		return s.lapseCardConf(card, deckConf)
	}
	return s.newCardConf(card, deckConf)
}

func parent(name string) string {
	parts := strings.Split(name, "::")
	if len(parts) < 2 {
		return ""
	}
	parts = parts[:len(parts)-1]
	return strings.Join(parts, "")
}

func updateDeck(deck *models.Deck, todayStmp int64) {
	if deck.NewToday[0] != todayStmp {
		deck.NewToday = [2]int64{todayStmp, 0}
	}
	if deck.ReviewsToday[0] != todayStmp {
		deck.ReviewsToday = [2]int64{todayStmp, 0}
	}

	if deck.LearnToday[0] != todayStmp {
		deck.LearnToday = [2]int64{todayStmp, 0}
	}
}

func (s schedV2Service) previewingCard(card models.Card) (bool, error) {
	conf, err := s.deckRepo.Conf(card.DeckID)
	if err != nil {
		return false, err
	}
	return bool(conf.Dyn) && !conf.Resched, nil
}

func (s schedV2Service) answerCardPreview(card models.Card, ease models.Ease, conf models.DeckConfig) (models.Card, error) {
	if ease == 1 {
		card.Queue = models.CardQueuePreview
		var previewDelay models.UnixTime
		if conf.PreviewDelay == nil {
			previewDelay = 10
		} else {
			previewDelay = *conf.PreviewDelay
		}
		card.Due = models.UnixTime(time.Now().Unix() + int64(previewDelay)*60)
		s.learningCount = s.learningCount + 1
		return card, nil
	} else {
		card = s.restorePreviewCard(card)
		card = s.removeFromFiltered(card)
	}
	return card, nil
}

func (s schedV2Service) answerLearnCard(card models.Card, deckConfig models.DeckConfig, ease models.Ease) error {
	conf, err := s.learnCardConf(card, deckConfig)
	if err != nil {
		return err
	}

	var delays []int64
	switch c := conf.(type) {
	case models.NewDeckConf:
		delays = c.Delays
	case models.LapseDeckConf:
		delays = c.Delays
	default:
		return fmt.Errorf("could not determine deck config type")
	}
	var revLogType models.ReviewLogType
	if card.Type == models.CardTypeReview || card.Type == models.CardTypeRelearning {
		revLogType = models.ReviewLogTypeRelearn
	} else {
		revLogType = models.ReviewLogTypeLearning
	}
	var leaving bool
	// lrnCount was decremented once when card was fetched
	lastLeft := card.ReviewsLeft
	if ease == models.ReviewEaseEasy {
		s.rescheduleAsReviewed(card, deckConfig, true)
		leaving = true
	} else if ease == models.ReviewEaseOK {
		if (card.ReviewsLeft%1000)-1 <= 0 {
			s.rescheduleAsReviewed(card, deckConfig, false)
			leaving = true
		} else {
			s.moveToNextStep(card, delays)
		}

	} else if ease == models.ReviewEaseHard {
		colConf, err := s.colRepo.Conf()
		if err != nil {
			return err
		}
		s.repeatStep(card, delays, colConf)
	} else {
    if _, err := s.moveToFirstStep(card, deckConfig, delays); err != nil {
      return err
    }
	}

	return s.logLearn(card, deckConfig, delays, lastLeft, leaving, ease, revLogType)
}

func (s schedV2Service) answerReviewCard(card models.Card, ease models.Ease, deckConf models.DeckConfig) error {
	var (
		early      bool
		delay      int64
		revLogType models.ReviewLogType
	)

	if card.OriginalDeckID != 0 && int64(card.OriginalDue) > s.today {
		early = true
	}
	if early {
		revLogType = models.ReviewLogTypeCram
	} else {
		revLogType = models.ReviewLogTypeReview
	}

	if ease == models.ReviewEaseWrong {
		var err error
		delay, err = s.rescheduleLapse(card, deckConf)
		if err != nil {
			return err
		}
	} else {
		s.rescheduleReview(card, ease, early, deckConf)
	}

	usn, err := s.colRepo.USN(true)
	if err != nil {
		return err
	}
	return s.revLogRepo.Create(card, usn, ease, -delay, card.LastInterval, TimeTaken(card, deckConf), revLogType)
}

func (s schedV2Service) restorePreviewCard(card models.Card) models.Card {
	previewCard := card
	previewCard.Due = card.OriginalDue
	if card.Type == models.CardTypeLearning || card.Type == models.CardTypeRelearning {
		if card.OriginalDue > 1000000000 {
			card.Queue = models.CardQueueLearning
		} else {
			card.Queue = models.CardQueueRelearning
		}
	} else {
		card.Queue = models.CardQue(card.Type)
	}
	return previewCard
}

func (s schedV2Service) removeFromFiltered(card models.Card) models.Card {
	if card.OriginalDeckID != 0 {
		card.Deck.ID = card.OriginalDeckID
		card.DeckID = card.OriginalDeckID
		card.OriginalDeckID = 0
		card.OriginalDue = 0
	}
	return card
}

func (s schedV2Service) startingLeft(card models.Card, deckConf models.DeckConfig) (int, error) {
	var delays []int64
	if card.Type == models.CardTypeRelearning {
		laspeConf, err := s.lapseCardConf(card, deckConf)
		delays = laspeConf.Delays
		if err != nil {
			return -1, err
		}
	} else {
		conf, err := s.learnCardConf(card, deckConf)
		if err != nil {
			return -1, err
		}
		// TODO: figure out a better way to determine the type of conf
		switch c := conf.(type) {
		case models.NewDeckConf:
			delays = c.Delays
		case models.LapseDeckConf:
			delays = c.Delays
		default:
			return -1, fmt.Errorf("could not determine deck config type")
		}
	}
	totalDelays := len(delays)
	return totalDelays + s.leftToday(delays, totalDelays)*1000, nil
}

// leftToday return the number of steps that can be completed by the day cutoff
func (s schedV2Service) leftToday(delays []int64, left int) int {
	now := time.Now().Unix()
	offset := sint.Min(left, len(delays))
	var steps int
	for i := 0; i < offset; i++ {
		now += int64(delays[i] * 60)
		if now > s.dayCutoff {
			break
		}
		steps = i
	}
	return steps + 1
}

func (s schedV2Service) updateStats(card models.Card, cardType models.CardType, timeCount int64) error {
	decks, err := s.deckRepo.DeckWithParents(card.DeckID)
	if err != nil {
		return err
	}
	for _, d := range decks {
		switch cardType {
		case models.CardTypeNew:
			d.NewToday[1] += timeCount
		case models.CardTypeLearning, models.CardTypeRelearning:
			d.LearnToday[1] += timeCount
		case models.CardTypeReview:
			d.ReviewsToday[1] += timeCount
		default:
			// add a log?
			// do nothing
		}
		saveErr := s.deckRepo.Save(&d)
		if saveErr != nil {
			return saveErr
		}
	}
	return nil
}
func (s schedV2Service) rescheduleGraduatingLapse(card models.Card, early bool) models.Card {
	if early {
		card.Interval += 1
	}
	card.Due = models.UnixTime(s.today) + models.UnixTime(card.Interval)
	card.Queue = models.CardQueueReview
	card.Type = models.CardTypeReview
	return card
}

func (s schedV2Service) fuzzIntervalRange(interval int64) (int64, int64) {
	var fuzz int64
	if interval < 2 {
		return 1, 1
	} else if interval == 2 {
		return 2, 3
	} else if interval < 7 {
		fuzz = int64(float64(interval) * 0.25)
	} else if interval < 30 {
		fuzz = utils.MaxInt64(2, int64(float64(interval)*0.15))
	} else {
		fuzz = utils.MaxInt64(4, (int64(float64(interval) * 0.05)))
	}
	// fuzz at least a day
	fuzz = utils.MaxInt64(fuzz, 1)
	return interval - fuzz, interval + fuzz
}

func (s schedV2Service) fuzzedInterval(interval int64) int64 {
	min, max := s.fuzzIntervalRange(interval)
	return rand.Int63n(max+1) + min
}

func (s schedV2Service) graduatingInterval(card models.Card, deckConfig models.DeckConfig, early bool, fuzzy bool) int64 {
	var bonus int64
	if card.Type == models.CardTypeReview || card.Type == models.CardTypeRelearning {
		bonus = 1
		if early {
			bonus = 0
		}
		return card.Interval + bonus
	}
	var ideal int64
	if !early {
		// graduate / complete
		ideal = deckConfig.New.Ints[0]
	} else {
		// early removal
		ideal = deckConfig.New.Ints[1]
	}
	if fuzzy {
		ideal = s.fuzzedInterval((ideal))
	}
	return ideal
}

// rescheduleNew will reschedule a new card that is graduated/completed for the first time
func (s schedV2Service) rescheduleNew(card models.Card, deckConfig models.DeckConfig, early bool) {
	card.Interval = s.graduatingInterval(card, deckConfig, early, true)
	card.Due = models.UnixTime(s.today + int64(card.Interval))
	card.Factor = deckConfig.New.InitialFactor
	card.Type = models.CardTypeReview
	card.Queue = models.CardQueueReview
}

func (s schedV2Service) rescheduleLapse(card models.Card, deckConf models.DeckConfig) (int64, error) {
	laspeConf, err := s.lapseCardConf(card, deckConf)
	card.Lapses += 1
	card.Factor = utils.MaxInt64(1300, card.Factor-200)
	var (
		suspended bool
		delay     int64
	)
	isLeechCard, err := s.isLeechCard(card, deckConf)
	if err != nil {
		return -1, err
	}
	if isLeechCard && card.Queue == models.CardQueueSuspended {
		suspended = true
	}
	if len(laspeConf.Delays) > 0 && !suspended {
		card.Type = models.CardTypeRelearning
		var stepErr error
		delay, stepErr = s.moveToFirstStep(card, deckConf, laspeConf.Delays)
		if stepErr != nil {
			return -1, stepErr
		}
	} else {
		// no relearning steps
		card.LastInterval = card.Interval
		card.Interval = utils.MaxOfInt64(1, int64(laspeConf.MinInterval), card.Interval*laspeConf.Mult)
		if err := s.rescheduleAsReviewed(card, deckConf, false); err != nil {
			return -1, err
		}

		if suspended {
			card.Queue = models.CardQueueSuspended
		}
	}
	return delay, nil
}
func (s schedV2Service) reviewCardConfig(card models.Card, deckConfig models.DeckConfig) (models.RevDeckConf, error) {
	// normal deck
	if card.OriginalDeckID != 0 {
		return deckConfig.Rev, nil
	}
	origConf, err := s.deckRepo.Conf(card.OriginalDeckID)
	if err != nil {
		return models.RevDeckConf{}, err
	}

	return origConf.Rev, nil
}

func (s schedV2Service) constrainInterval(interval int64, revConf models.RevDeckConf, previous int64, fuzz bool) int64 {
	var intervalFct int64
	if revConf.IvlFct != nil {
		intervalFct = *revConf.IvlFct
	} else {
		intervalFct = 1.0
	}
	intervalNew := int64(interval * intervalFct)
	if fuzz {
		intervalNew = s.fuzzedInterval(intervalNew)
	}
	intervalNew = utils.MaxOfInt64(intervalNew, previous+1, 1)
	intervalNew = utils.MinInt64(intervalNew, revConf.MaxIvl)
	return intervalNew
}

func (s schedV2Service) earlyReviewInterval(card models.Card, ease models.Ease, deckConf models.DeckConfig) (int64, error) {
	elapsed := int64(card.Interval) - int64(card.OriginalDue) - s.today
	conf, err := s.reviewCardConfig(card, deckConf)
	if err != nil {
		return 0, err
	}
	var easyBonus float64 = 1
	// early 3/4 reviews shouldn't decrease previous interval
	var minNewInterval float64 = 1
	var (
		factor     float64
		hardFactor float64
	)
	if conf.HardFactor == nil {
		hardFactor = 1.2
	} else {
		hardFactor = *conf.HardFactor
	}

	if ease == models.ReviewEaseHard {
		factor = hardFactor
		// hard cards shouldn't have their interval decreased by more than 50%
		// of the normal factor
		minNewInterval = factor / 2
	} else if ease == models.ReviewEaseOK {
		factor = float64(card.Factor) / 1000
	} else {
		// ease == 4
		factor = float64(card.Factor) / 1000
		ease4 := conf.Ease4
		// 1.3 -> 1.15
		easyBonus = ease4 - (ease4-1)/2
	}
	interval := utils.MaxOfInt64(elapsed*int64(factor), 1.0)
	// cap interval decreases
	interval = utils.MaxOfInt64(card.Interval*int64(minNewInterval), interval) * int64(easyBonus)
	interval = s.constrainInterval(interval, conf, 0, false)

	return interval, nil
}

// TODO: check if method is shared with schedv1
func (s schedV2Service) daysLate(card models.Card) int64 {
	// "Number of days later than scheduled."
	var due int64
	if card.OriginalDeckID != 0 {
		due = int64(card.OriginalDeckID)
	} else {
		due = int64(card.Due)
	}
	return utils.MaxInt64(0, s.today-due)
}

// TODO: check if method is shared with schedv1
func (s schedV2Service) nextReviewInterval(card models.Card, ease models.Ease, fuzz bool, deckConfig models.DeckConfig) (int64, error) {
	//"Next review interval for CARD, given EASE."
	delay := s.daysLate(card)
	conf, err := s.reviewCardConfig(card, deckConfig)
	if err != nil {
		return 0, err
	}
	factor := card.Factor / 1000
	var hardFactor float64

	if conf.HardFactor == nil {
		hardFactor = 1.2
	} else {
		hardFactor = *conf.HardFactor
	}

	var hardMin int64
	if hardFactor > 1 {
		hardMin = int64(card.Interval)
	}

	ivl2 := card.Interval * int64(hardFactor)
	ivl := s.constrainInterval(ivl2, conf, hardMin, fuzz)
	if ease == models.LearnEaseOK {
		return ivl, nil
	}

	ivl3 := (card.Interval + delay) / 2 * int64(factor)
	ivl = s.constrainInterval(ivl3, conf, ivl, fuzz)
	if ease == models.LearnEaseEasy {
		return ivl, nil
	}
	ivl4 := (card.Interval + delay) * int64(hardFactor) * int64(conf.Ease4)
	ivl = s.constrainInterval(ivl4, conf, ivl, fuzz)
	return ivl, nil
}

func (s schedV2Service) rescheduleReview(card models.Card, ease models.Ease, early bool, deckConfig models.DeckConfig) error {
	// update interval
	card.LastInterval = card.Interval
	var (
		err error
		ivl int64
	)
	if early {
		ivl, err = s.earlyReviewInterval(card, ease, deckConfig)
	} else {
		ivl, err = s.nextReviewInterval(card, ease, true, deckConfig)
	}

	if err != nil {
		return err
	}
	card.Interval = ivl

	//then the rest
	card.Factor = utils.MaxInt64(1300, card.Factor+FactorAdditionValues[ease-2])
	card.Due = models.UnixTime(s.today + card.Interval)

	// card leaves filtered deck
	s.removeFromFiltered(card)
	return nil
}

func (s schedV2Service) rescheduleAsReviewed(card models.Card, deckConfig models.DeckConfig, early bool) error {
	if card.Type == models.CardTypeReview || card.Type == models.CardTypeRelearning {
		card = s.rescheduleGraduatingLapse(card, early)
	} else {
		s.rescheduleNew(card, deckConfig, early)
	}

	if card.OriginalDeckID != 0 {
		s.removeFromFiltered(card)
	}
	return nil
}

func (s schedV2Service) moveToFirstStep(card models.Card, conf models.DeckConfig, delays []int64) (int64, error) {
	var err error
	colConf, err := s.colRepo.Conf()
	if err != nil {
		return -1, err
	}
	card.ReviewsLeft, err = s.startingLeft(card, conf)
	if err != nil {
		return -1, err
	}
	// relearning card?
	if card.Type == models.CardTypeRelearning {
		s.updateReviewIntervalOnFail(card, conf)
	}
	return s.rescheduleLrnCard(card, delays, nil, colConf), nil
}

// moveToNextStep determines how many card left to study by decrementing true remaining count from today
func (s schedV2Service) moveToNextStep(card models.Card, delays []int64) error {
	colConf, err := s.colRepo.Conf()
	if err != nil {
		return err
	}
	left := (card.ReviewsLeft % 1000) - 1
	card.ReviewsLeft = s.leftToday(delays, left)*1000 + left
	s.rescheduleLrnCard(card, delays, nil, colConf)
	return nil
}

func (s schedV2Service) delayForGrade(delays []int64, left int) (delay int64) {
	left = left % 1000
	if len(delays)-left > -1 {
		delay = delays[len(delays)-left]
	} else if len(delays) > 0 {
		delay = delays[0]
	} else {
		// use dummy value
		delay = 1
	}
	delay = delay * 60
  return
}

func (s schedV2Service) rescheduleLrnCard(card models.Card, delays []int64, delay *int64, colConf models.CollectionConf) int64 {
	if delay == nil {
    delayGrade := s.delayForGrade(delays, card.ReviewsLeft)
		delay = &delayGrade
	}
	card.Due = models.UnixTime(time.Now().Unix() + int64(*delay))
	if card.Due < models.UnixTime(s.dayCutoff) {
		maxExtra := math.Min(300, float64(*delay)*0.25)
		fuzz := rand.Int63n(int64(math.Max(1, maxExtra)))
		card.Due = models.UnixTime(math.Min(float64(s.dayCutoff)-1, float64(int64(card.Due)+fuzz)))
		card.Queue = models.CardQueueLearning
		if card.Due < (models.UnixTime(time.Now().Unix() + colConf.CollapseTime)) {
			s.learningCount = s.learningCount + 1
			// prevent adding entry into queue twice by checking if queue is not empty
			if s.lrnQueue != nil && s.revCount < 1 && s.newCount < 1 {
				smallestDue := int(s.lrnQueue[0].Due)
				card.Due = models.UnixTime(math.Max(float64(int64(card.Due)), float64((smallestDue)+1)))
			}
			s.lrnQueue = append(s.lrnQueue, &learnQueue{
				Due: card.Due,
				ID:  card.ID,
			})
			sort.Sort(ByLearnDue(s.lrnQueue))
		}
	} else {
		ahead := math.Floor(float64(int64(card.Due)-s.dayCutoff)/float64(86400)) + 1
		card.Due = models.UnixTime(int64(s.today) + int64(ahead))
		card.Queue = models.CardQueueRelearning
	}
	return *delay
}

func (s schedV2Service) delayForRepeatingGrade(left int, delays []int64) int64 {
	var delay2 int64
	// halfway between last and next
	delay1 := s.delayForGrade(delays, left)
	if len(delays) > 1 {
		delay2 = s.delayForGrade(delays, left-1)
	} else {
		delay2 = delay1 * 2
	}
	avg := (delay1 + utils.MaxInt64(delay1, delay2)) // 2
	return avg
}

func (s schedV2Service) logLearn(card models.Card, conf models.DeckConfig, delays []int64, left int, leaving bool, ease models.Ease, revLogType models.ReviewLogType) error {
	lastInterval := -(s.delayForGrade(delays, left))
	var interval int64
	if leaving {
		interval = card.Interval
	} else if ease == models.LearnEaseOK {
		interval = -(s.delayForRepeatingGrade(card.ReviewsLeft, delays))
	} else {
		interval = -(s.delayForGrade(delays, card.ReviewsLeft))
	}

	usn, err := s.colRepo.USN(true)
	if err != nil {
		return err
	}
	return s.revLogRepo.Create(card, usn, ease, interval, lastInterval, TimeTaken(card, conf), revLogType)
}

func (s schedV2Service) repeatStep(card models.Card, delays []int64, colConf models.CollectionConf) {
	delay := s.delayForRepeatingGrade(card.ReviewsLeft, delays)
	s.rescheduleLrnCard(card, delays, &delay, colConf)
}

func (s schedV2Service) updateReviewIntervalOnFail(card models.Card, conf models.DeckConfig) {
	card.LastInterval = card.Interval
	card.Interval = utils.MaxInt64(conf.Lapse.MinInterval, card.Interval*conf.Lapse.Mult)
}

// TimeTaken returns the time taken to answer card, in integer MS."
func TimeTaken(card models.Card, conf models.DeckConfig) int64 {
	total := time.Now().Unix() - int64(card.TimeStarted)*1000
	return utils.MaxInt64(total, conf.MaxTaken)
}

func (s schedV2Service) isLeechCard(card models.Card, conf models.DeckConfig) (bool, error) {
	leechFails := conf.Lapse.LeechFails
	if leechFails == 0 {
		return false, nil
	}

	lf := leechFails / 2
	if card.Lapses >= leechFails && (card.Lapses-leechFails)%(sint.Max(int(lf), 1)) == 0 {
		note := card.Note
		note.Model.Tags = append(note.Model.Tags, "leech")
		if err := s.noteRepo.Create(note); err != nil {
			return false, err
		}
		if conf.Lapse.LeechAction == 0 {
			card.Queue = models.CardQueueSuspended
		}
		return true, nil
	}
	return false, nil
}

func (s schedV2Service) NextLearnInterval(card models.Card, ease models.Ease, deckConf models.DeckConfig) (int64, error) {
	if card.Queue == models.CardQueueNew {
		left, err := s.startingLeft(card, deckConf)
		if err != nil {
			return 0, err
		}
		card.ReviewsLeft = left
	}
	conf, err := s.learnCardConf(card, deckConf)
	if err != nil {
		return 0, err
	}

	var delays []int64
	switch c := conf.(type) {
	case models.NewDeckConf:
		delays = c.Delays
	case models.LapseDeckConf:
		delays = c.Delays
	default:
		return 0, fmt.Errorf("could not determine deck config type")
	}

	if ease == models.LearnEaseWrong {
		return s.delayForGrade(delays, card.ReviewsLeft), nil
	} else if ease == models.LearnEaseOK {
		return s.delayForRepeatingGrade(card.ReviewsLeft, delays), nil
	} else if ease == models.ReviewEaseEasy {
		return s.graduatingInterval(card, deckConf, true, false) * 86400, nil
	} else {
		reviewsLeft := card.ReviewsLeft%1000 - 1
		if reviewsLeft <= 0 {
			return s.graduatingInterval(card, deckConf, false, false), nil
		} else {
			return s.delayForGrade(delays, reviewsLeft), nil
		}
	}
}

// NextIntervalString returns the next interval for CARD as a string.
func (s schedV2Service) NextIntervalString(card models.Card, ease models.Ease, conf models.DeckConfig) (string, error) {
	colConf, err := s.colRepo.Conf()
	if err != nil {
		return "", err
	}
	ivl, err := s.NextInterval(card, ease, conf)
	if err != nil {
		return "", err
	}
	if ivl == 0 {
		return "(end)", nil
	}
	ivlStr, err := utils.FormatTimeSpan(ivl, 0, 0, false, false, nil)
	if err != nil {
		return "", err
	}
	if ivl < colConf.CollapseTime {
		ivlStr = "<" + ivlStr
	}

	return ivlStr, nil
}

// NextInterval returns the next interval for CARD, in seconds
func (s schedV2Service) NextInterval(card models.Card, ease models.Ease, conf models.DeckConfig) (int64, error) {
	// preview mode
	previewCard, err := s.previewingCard(card)
	if err != nil {
		return 0, err
	}
	if previewCard {
		if ease == models.ReviewEaseWrong {
			val, err := conf.Get("previewDelay", 10)
			if err != nil {
				return 0, err
			}
			return val.Int() * 60, nil
		}
		return 0, nil
	}
	// (re)learning?
	if card.Queue == models.CardQueueNew ||
		card.Queue == models.CardQueueLearning ||
		card.Queue == models.CardQueueRelearning {
		return s.NextLearnInterval(card, ease, conf)
	} else if ease == models.ReviewEaseWrong {
		// lapse
		laspeConf, err := s.lapseCardConf(card, conf)
		if err != nil {
			return 0, err
		}
		if len(laspeConf.Delays) > 0 {
			return laspeConf.Delays[0] * 60, nil
		}
		ivl := utils.MaxOfInt64(1, laspeConf.MinInterval, card.Interval*laspeConf.Mult)
		return ivl, nil
	} else {
		// review
		if card.OriginalDeckID > 0 && card.OriginalDue > models.UnixTime(s.today) {
			return s.earlyReviewInterval(card, ease, conf)
		} else {
			return s.nextReviewInterval(card, ease, false, conf)
		}
	}
}
