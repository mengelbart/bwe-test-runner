package common

type Config struct {
	Date        int64        `json:"date"`
	Connections []Connection `json:"connections"`
	Scenario    Scenario     `json:"scenario"`
}

type Connection struct {
	Name           string `json:"name"`
	Router         string `json:"router"`
	Implementation string `json:"implementation"`
}

type Scenario struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}
