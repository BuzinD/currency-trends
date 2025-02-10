package store

import (
	"cur/internal/model"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type CandleRepository struct {
	db *sql.DB
}

func NewCandleRepository(db *sql.DB) *CandleRepository {
	return &CandleRepository{
		db: db,
	}
}

func (rep *CandleRepository) InsertCandles(candles *[]model.Candle) error {
	query := strings.Join([]string{"INSERT INTO candles (pair, timestamp, open_price, high_price, low_price, close_price, volume, bar)",
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		"ON CONFLICT (pair, timestamp, bar)",
		"DO UPDATE SET open_price = EXCLUDED.open_price,",
		"high_price = EXCLUDED.high_price,",
		"low_price = EXCLUDED.low_price,",
		"close_price = EXCLUDED.close_price,", "" +
			"volume = EXCLUDED.volume;",
	},
		" ")

	tx, err := rep.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, candle := range *candles {
		_, err := tx.Exec(query, candle.Pair, candle.Timestamp, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume, candle.Bar)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return fmt.Errorf("failed to insert/update candles: %w", err)
			}
			return fmt.Errorf("failed to insert/update candles: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (rep *CandleRepository) FetchAll() ([]model.Candle, error) {
	query := "SELECT * FROM candles"

	rows, err := rep.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToCandles(rows)
}

func rowsToCandles(rows *sql.Rows) ([]model.Candle, error) {
	var candles []model.Candle
	for rows.Next() {
		var candle model.Candle
		err := rows.Scan(&candle.Pair, &candle.Timestamp, &candle.Open, &candle.High, &candle.Low, &candle.Close, &candle.Volume, &candle.Bar)
		if err != nil {
			return nil, err
		}
		candles = append(candles, candle)
	}

	return candles, nil
}

func (rep *CandleRepository) fetchByCurrency(pair string, from, to time.Time) ([]model.Candle, error) {
	query := "SELECT * FROM candles WHERE pair=$1 and `timestamp` BETWEEN($2,$3)"

	rows, err := rep.db.Query(query, pair, from.Unix(), to.Unix())

	if err != nil {
		return nil, err
	}
	defer rows.Close() // Ensure rows are closed after function execution

	return rowsToCandles(rows)
}

// GetLastTsForPair getting max timestamp in milliseconds
func (rep *CandleRepository) GetLastTsForPair(pair string) (string, error) {
	query := "SELECT (EXTRACT(EPOCH FROM timestamp) * 1000)::BIGINT::TEXT as ts  FROM candles WHERE pair=$1 ORDER BY timestamp DESC LIMIT 1"
	var lastTimestamp string
	err := rep.db.QueryRow(query, pair).Scan(&lastTimestamp)
	if err != nil {
		return "", err
	}
	return lastTimestamp, nil
}

// GetFirstTsForPair getting max timestamp in milliseconds
func (rep *CandleRepository) GetFirstTsForPair(pair string) (string, error) {
	query := "SELECT (EXTRACT(EPOCH FROM timestamp) * 1000)::BIGINT::TEXT as ts  FROM candles WHERE pair=$1 ORDER BY timestamp ASC LIMIT 1"
	var lastTimestamp string
	err := rep.db.QueryRow(query, pair).Scan(&lastTimestamp)
	if err != nil {
		return "", err
	}
	return lastTimestamp, nil
}
