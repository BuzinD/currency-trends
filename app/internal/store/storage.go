package store

import (
	"database/sql"
)

type Store struct {
	db          *sql.DB
	currencyRep *CurrencyRepository
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Currency() *CurrencyRepository {
	if s.currencyRep != nil {
		return s.currencyRep
	}

	s.currencyRep = NewCurrencyRepository(s.db)

	return s.currencyRep
}
