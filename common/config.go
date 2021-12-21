package common

type Config struct {
	Date            int64            `json:"date"`
	Implementations []Implementation `json:"implementations"`
	Scenario        Scenario         `json:"scenario"`
}

type Implementation struct {
	Name   string `json:"name"`
	Router string `json:"router"`
	Source string `json:"source"`
}

type Scenario struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}
