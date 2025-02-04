package model

type Currency struct {
	Id          int64
	Code        string
	Chain       string
	CanDeposit  bool
	CanWithdraw bool
}
