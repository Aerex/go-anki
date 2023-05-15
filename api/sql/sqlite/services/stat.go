package services

import (
	repos "github.com/aerex/go-anki/api/sql/sqlite/repositories"
)

type StatService struct {
	revLogRepo repos.RevLogRepo
}
