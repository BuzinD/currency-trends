package store

import (
	"cur/internal/model"
	"cur/internal/service/okx/response"
	"database/sql"
	"fmt"
	"strings"
)

type CurrencyRepository struct {
	db *sql.DB
}

func NewCurrencyRepository(db *sql.DB) *CurrencyRepository {
	return &CurrencyRepository{
		db: db,
	}
}

func (rep *CurrencyRepository) InsertOrUpdateCurrencies(currencies *[]response.CurrencyResponseData) error {
	query := strings.Join([]string{"INSERT INTO currencies (code, chain, can_deposit, can_withdraw)	VALUES ($1, $2, $3, $4)",
		"ON CONFLICT (code, chain)",
		"DO UPDATE SET can_deposit = $3, can_withdraw = $4;"}, " ")

	tx, err := rep.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, currency := range *currencies {
		_, err := tx.Exec(query, currency.Ccy, currency.Chain, currency.CanDep, currency.CanWd)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert/update currency: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (rep *CurrencyRepository) FetchAll() ([]model.Currency, error) {
	query := "SELECT * FROM currencies"

	rows, err := rep.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Ensure rows are closed after function execution

	var currencies []model.Currency
	for rows.Next() {
		var currency model.Currency
		err := rows.Scan(&currency.Id, &currency.Code, &currency.Chain, &currency.CanDeposit, &currency.CanWithdraw)
		if err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return currencies, nil
}
