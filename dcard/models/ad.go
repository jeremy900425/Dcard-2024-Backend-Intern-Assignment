package models

import "time"

// Ad represents an advertisement with conditions
type Ad struct {
	Title      string    `json:"title"`
	StartAt    time.Time `json:"startAt"`
	EndAt      time.Time `json:"endAt"`
	Conditions struct {
		AgeStart int      `json:"ageStart,omitempty"`
		AgeEnd   int      `json:"ageEnd,omitempty"`
		Gender   string   `json:"gender,omitempty"`
		Country  []string `json:"country,omitempty"`
		Platform []string `json:"platform,omitempty"`
	} `json:"conditions"`
}
