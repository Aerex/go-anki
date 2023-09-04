package services

import (
	repos "github.com/aerex/go-anki/api/sql/sqlite/repositories"
	"github.com/aerex/go-anki/pkg/models"
)

type StatService struct {
	revLogRepo repos.RevLogRepo
	colRepo    repos.ColRepo
}

func NewStatsService(revlog repos.RevLogRepo, col repos.ColRepo) StatService {
	return StatService{
		revLogRepo: revlog,
		colRepo:    col,
	}
}

func (s *StatService) TodayStats() (models.StudiedToday, error) {
	cutoff := s.colRepo.DayCutoff()
	return s.revLogRepo.TodayStats(cutoff)
}

func (s *StatService) MaturedStats() (models.MaturedToday, error) {
	cutoff := s.colRepo.DayCutoff()
	return s.revLogRepo.MaturedCards(cutoff)
}
