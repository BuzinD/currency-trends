package structure

type TickerResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		InstId    string `json:"instId"`
		Last      string `json:"last"`
		Open24H   string `json:"open24h"`
		High24H   string `json:"high24h"`
		Low24H    string `json:"low24h"`
		Vol24H    string `json:"vol24h"`
		Timestamp string `json:"ts"`
	} `json:"data"`
}
