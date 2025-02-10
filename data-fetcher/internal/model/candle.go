package model

import "time"

type Candle struct {
	Pair      string
	Timestamp time.Time
	Open      int64
	High      int64
	Low       int64
	Close     int64
	Volume    int64
	Bar       string
}
