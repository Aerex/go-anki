package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/exp/maps"

	"github.com/google/gapid/core/math/sint"
	"github.com/op/go-logging"

	"github.com/aerex/go-anki/api/sql"
	repos "github.com/aerex/go-anki/api/sql/sqlite/repositories"
	"github.com/aerex/go-anki/pkg/models"
)

type learnQueue struct {
	Due models.UnixTime
	ID  models.ID
}

type SchedV2Service struct {
	colRepo        repos.ColRepo
	deckRepo       repos.DeckRepo
	cardsRepo      repos.CardRepo
	colConf        *models.CollectionConf
	server         bool
	revCount       int
	revQueue       []int
	newCount       int
	newCardModulus int
	dayCutoff      int64
	lrnCutoff      int64
	lrnQueue       []learnQueue
	lrnDayQueue    []int
	lrnDeckIDs     []models.ID
	newDeckIDs     []models.ID
	newQueue       []int
	haveQueues     bool
	learningCount  int
	reportLimit    int
	today          uint32
}

var logger = logging.MustGetLogger("ankicli")

const (
	DynReportLimit = 99999
	ReportLimit    = 1000
)

func NewSchedService(c repos.ColRepo, cd repos.CardRepo, d repos.DeckRepo, server bool) SchedV2Service {
	return SchedV2Service{
		colRepo:   c,
		deckRepo:  d,
		cardsRepo: cd,
		server:    server,
	}
}

func (s *SchedV2Service) DeckStudyStats() (map[models.ID]models.DeckStudyStats, error) {
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
		due := time.Now().Unix() + int64(s.colConf.CollapseTime)
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
func (s *SchedV2Service) checkDay() error {
	cutoff := s.colRepo.DayCutoff()
	if time.Now().Unix() > cutoff {
		if err := s.reset(false); err != nil {
			return err
		}
	}
	return nil
}

func (s *SchedV2Service) reset(server bool) error {
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

func (s *SchedV2Service) currentTimezoneOffset() (int32, error) {
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

func daysElapsed(startDate time.Time, endDate time.Time, rolloverPassed bool) int {
	days := (endDate.Sub(startDate).Abs().Hours()) / 24

	if rolloverPassed {
		return int(days)
	}
	return int(days - 1)
}

func (s *SchedV2Service) timingToday(crt models.UnixTime, crtMinWest int32, nowSec int64, nowMinWest int32, rolloverHr int) models.SchedTimingToday {
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
		DaysElapsed: uint32(daysElapsed),
		NextDayAt:   nextDateAt,
	}
}
func (s *SchedV2Service) _dayCutoff() int64 {
	rollover := s.colConf.Rollover
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

func (s *SchedV2Service) daysSinceCreation(crt models.UnixTime, rollover int) int64 {
	start := time.Unix(int64(crt), 0)
	start = time.Date(start.Year(), start.Month(), start.Day(), rollover, 0, 0, 0, time.UTC)
	return int64((time.Now().Unix() - start.Unix()) / 86400)
}

func (s *SchedV2Service) updateCutoff() error {
	var conf models.CollectionConf
	if s.colConf == nil {
		var err error
		conf, err = s.colRepo.Conf()
		if err != nil {
			return err
		}
		s.colConf = &conf
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
	timing := s.timingToday(createdTime, conf.CreationOffset, time.Now().Unix(), offset, int(conf.Rollover))
	if s.colConf.CreationOffset != 0 {
		s.today = timing.DaysElapsed
		s.dayCutoff = timing.NextDayAt
	} else {
		s.today = uint32(s.daysSinceCreation(createdTime, int(conf.Rollover)))
		s.dayCutoff = s._dayCutoff()
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
	if s.colConf.LastUnburied < s.today {
		if err := s.cardsRepo.UnburyCards(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SchedV2Service) resetRev() error {
	if err := s.resetRevCount(); err != nil {
		return err
	}
	s.revQueue = []int{}
	return nil
}

func (s *SchedV2Service) resetRevCount() error {
	limit, err := s.currentRevLimit()
	if err != nil {
		return err
	}

	deckLimit := sql.InClauseFromIDs(s.colConf.ActiveDecks)
	revisions, err := s.cardsRepo.Revisions(deckLimit, limit)
	if err != nil {
		return err
	}
	s.revCount = revisions

	return nil
}

func (s *SchedV2Service) currentRevLimit() (int, error) {
	decks, err := s.deckRepo.Decks()
	if err != nil {
		return 0, err
	}
	selectedDeck, exists := decks[models.ID(s.colConf.CurrentDeck)]
	if !exists {
		// TODO: log deck does not exist
		return 0, fmt.Errorf("deck %s could not be found", s.colConf.CurrentDeck)
	}
	return s.deckRevLimit(*selectedDeck, -1), nil
}

func (s *SchedV2Service) deckRevLimit(deck models.Deck, parentLimit int) int {
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
	} else {
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

	return 0
}

// deckLimitForNewCards get the limit for deck without parent limits
func (s *SchedV2Service) deckLimitForNewCards(deck models.Deck) (int, error) {
	if deck.Dyn {
		return DynReportLimit, nil
	}
	conf, err := s.deckRepo.Conf(models.ID(deck.Conf))
	if err != nil {
		return 0, err
	}
	return sint.Max(0, conf.New.PerDay-int(deck.NewToday[1])), nil
}

func (s *SchedV2Service) deckLimitForReviewCards(deck models.Deck, parentLimit int) (int, error) {
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
	} else {
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
	}
	return limit, nil
}

func (s *SchedV2Service) resetNew() error {
	newCount, err := s.resetNewCount()
	if err != nil {
		return err
	}
	s.newCount = newCount
	s.newDeckIDs = make([]models.ID, len(s.colConf.ActiveDecks))
	copy(s.newDeckIDs, s.colConf.ActiveDecks)
	s.newQueue = []int{}
	s.updateNewCardRatio()
	return nil
}

func (s *SchedV2Service) resetNewCount() (int, error) {
	return s.computeCount(s.deckLimitForNewCards, s.cardsRepo.CardsNewForDeck)
}

func (s *SchedV2Service) computeCount(lmtCb func(deck models.Deck) (int, error), cntCb func(deckID models.ID, limit int) (int, error)) (count int, err error) {
	decks, err := s.deckRepo.Decks()
	if err != nil {
		return 0, err
	}
	uniqueParentCounts := make(map[models.ID]int)
	for _, id := range s.colConf.ActiveDecks {
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

func (s *SchedV2Service) resetLrn() error {
	s.updateLrnCutoff(true)
	s.resetLrnCount()
	s.lrnQueue = []learnQueue{}
	s.lrnDayQueue = []int{}
	s.lrnDeckIDs = s.colConf.ActiveDecks
	return nil
}

func (s *SchedV2Service) updateLrnCutoff(force bool) bool {
	nextCutoff := time.Now().Unix() + int64(s.colConf.CollapseTime)
	if ((nextCutoff - s.lrnCutoff) > 60) || force {
		s.lrnCutoff = nextCutoff
		return true
	}
	return false
}

func (s *SchedV2Service) resetLrnCount() error {
	deckLimit := sql.InClauseFromIDs(s.colConf.ActiveDecks)

	learningCount, err := s.cardsRepo.LearningCount(deckLimit, s.lrnCutoff)
	if err != nil {
		return err
	}
	s.learningCount = learningCount
	return nil
}

func (s *SchedV2Service) updateNewCardRatio() {
	if s.colConf.NewSpread == models.NewCardSpread(models.NewCardsDistribute) {
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

func parent(name string) string {
	parts := strings.Split(name, "::")
	if len(parts) < 2 {
		return ""
	}
	parts = parts[:len(parts)-1]
	return strings.Join(parts, "")
}

func updateDeck(deck *models.Deck, todayStmp uint32) {
	if deck.NewToday[0] != todayStmp {
		deck.NewToday = [2]uint32{todayStmp, 0}
	}
	if deck.ReviewsToday[0] != todayStmp {
		deck.ReviewsToday = [2]uint32{todayStmp, 0}
	}

	if deck.LearnToday[0] != todayStmp {
		deck.LearnToday = [2]uint32{todayStmp, 0}
	}
}
