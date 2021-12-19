package common

type Config struct {
	Date           int64          `json:"date"`
	Implementation Implementation `json:"implementation"`
	Scenario       Scenario       `json:"scenario"`
}

type Implementation struct {
	Name string `json:"name"`
}

type Scenario struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}
