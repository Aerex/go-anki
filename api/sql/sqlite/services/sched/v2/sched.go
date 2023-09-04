package services

import (
	"time"

	"github.com/google/gapid/core/math/sint"
	"github.com/op/go-logging"

	repos "github.com/aerex/go-anki/api/sql/sqlite/repositories"
	"github.com/aerex/go-anki/pkg/models"
)

type schedService struct {
	colRepo   repos.ColRepo
	deckRepo  repos.DeckRepo
	cardsRepo repos.CardRepo
	colConf   *models.CollectionConf
	server    bool
	dayCutoff int64
	today     uint32
}

var logger = logging.MustGetLogger("ankicli")

func NewSchedService(c repos.ColRepo, d repos.DeckRepo, server bool) schedService {
	return schedService{
		colRepo:  c,
		deckRepo: d,
		server:   server,
	}
}

// CheckDay will check if the day has rolled over
// passed the cutoff day. If so, reset
func (s *schedService) CheckDay(server bool) {
	cutoff := s.colRepo.DayCutoff()
	if time.Now().Unix() > cutoff {
		s.reset()

	}

	//# check if the day has rolled over
	//if time.time() > self.dayCutoff:
	//    self.reset()
}

func (s *schedService) reset(server bool) {
	s.updateCutoff()
}

func (s *schedService) currentTimezoneOffset() (int32, error) {
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

func (s *schedService) timingToday(crt models.UnixTime, crtMinWest int32, nowSec int64, nowMinWest int32, rolloverHr int) models.SchedTimingToday {
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
func (s *schedService) _dayCutoff() int64 {
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

func (s *schedService) daysSinceCreation(crt models.UnixTime, rollover int) int64 {
	start := time.Unix(int64(crt), 0)
	start = time.Date(start.Year(), start.Month(), start.Day(), rollover, 0, 0, 0, time.UTC)
	return int64((time.Now().Unix() - start.Unix()) / 86400)
}

func (s *schedService) updateCutoff() error {
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
	now := time.Now().Unix()
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
		unburyCards()
	}
	//        unburied = self.col.conf.get("lastUnburied", 0)
	//        if unburied < self.today:
	//            self.unburyCards()
	//            self.col.conf["lastUnburied"] = self.today

	//   timing = self._timing_today()

	//   if self._new_timezone_enabled():
	//       self.today = timing.days_elapsed
	//       self.dayCutoff = timing.next_day_at
	//   else:
	//       self.today = self._daysSinceCreation()
	//       self.dayCutoff = self._dayCutoff()

	//   if oldToday != self.today:
	//       self.col.log(self.today, self.dayCutoff)

	//   # update all daily counts, but don't save decks to prevent needless
	//   # conflicts. we'll save on card answer instead
	//   def update(g):
	//       for t in "new", "rev", "lrn", "time":
	//           key = t + "Today"
	//           if g[key][0] != self.today:
	//               g[key] = [self.today, 0]

	//   for deck in self.col.decks.all():
	//       update(deck)
	//   # unbury if the day has rolled over
	//   unburied = self.col.conf.get("lastUnburied", 0)
	//   if unburied < self.today:
	//       self.unburyCards()
	//       self.col.conf["lastUnburied"] = self.today
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

func (s *schedService) unburyCards() {
	s.cardsRepo.Un

}
