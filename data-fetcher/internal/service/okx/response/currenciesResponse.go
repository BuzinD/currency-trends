package response

type CurrencyResponse struct {
	Code string                 `json:"code"`
	Msg  string                 `json:"msg"`
	Data []CurrencyResponseData `json:"data"`
}

type CurrencyResponseData struct {
	Ccy    string `json:"ccy"`
	Chain  string `json:"chain"`
	CanDep bool   `json:"canDep"`
	CanWd  bool   `json:"canWd"`
}
