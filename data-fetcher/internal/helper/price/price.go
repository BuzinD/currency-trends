package price

import (
	"errors"
	"strconv"
)

const PriceFactor = 100_000_000
const AdditionalZeroes = 8

type Price struct {
	Price int64
}

func (p Price) Add(other Price) Price {
	return Price{Price: p.Price + other.Price}
}

func (p Price) Sub(other Price) Price {
	return Price{Price: p.Price - other.Price}
}

func (p Price) Mul(factor int64) Price {
	return Price{Price: p.Price * factor}
}

func (p Price) Div(divisor int64) (Price, error) {
	if divisor == 0 {
		return Price{0}, errors.New("division by zero")
	}
	return Price{Price: p.Price / divisor}, nil
}

func (p Price) ToFloat() float64 {
	return float64(p.Price) / float64(PriceFactor)
}

// ParsePrice returns price in int64
func ParsePrice(priceStr string) (int64, error) {
	priceFloat, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, err
	}
	return int64(priceFloat * PriceFactor), nil
}
