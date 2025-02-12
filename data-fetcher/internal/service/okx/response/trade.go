package response

type TradeMessage struct {
	Arg struct {
		Channel string `json:"channel"`
		InstId  string `json:"instId"`
	} `json:"arg"`
	Data []struct {
		TradeID string `json:"tradeId"`
		Price   string `json:"px"`
		Size    string `json:"sz"`
		Side    string `json:"side"`
		Time    string `json:"ts"`
	} `json:"data"`
}
