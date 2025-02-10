package store

import (
	"database/sql"
	"strings"
)

type Store struct {
	db          *sql.DB
	currencyRep *CurrencyRepository
	candleRep   *CandleRepository
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Currency() *CurrencyRepository {
	if s.currencyRep == nil {
		s.currencyRep = NewCurrencyRepository(s.db)
	}

	return s.currencyRep
}

func (s *Store) Candle() *CandleRepository {
	if s.candleRep == nil {
		s.candleRep = NewCandleRepository(s.db)
	}

	return s.candleRep
}

func (s *Store) TruncateTables(tables []string) error {
	if len(tables) > 0 {
		_, err := s.db.Exec("TRUNCATE " + strings.Join(tables, ",") + " CASCADE")
		return err
	}
	return nil
}

func (s *Store) CloseConnection() {
	if s.db != nil {
		err := s.db.Close()
		if err != nil {
			return
		}
	}
}
