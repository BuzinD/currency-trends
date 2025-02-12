package request

type Arg struct {
	Channel string `json:"channel"`
	InstId  string `json:"instId"`
}

type SubscriptionMessage struct {
	Op   string `json:"op"`
	Args []Arg  `json:"args"`
}
