package alerter

type Alerts map[string][]Alert

type Alert struct {
	Message              string `json:"message"`
	ShouldIgnoreResource bool   `json:"-"`
}
