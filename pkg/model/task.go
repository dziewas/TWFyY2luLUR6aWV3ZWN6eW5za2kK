package model

type Task struct {
	Id       int    `json:"id,omitempty"`
	Url      string `json:"url,omitempty"`
	Interval int    `json:"interval,omitempty"`
}

type Attempt struct {
	Response  string  `json:"response,omitempty"`
	CreatedAt int64   `json:"created_at,omitempty"`
	Duration  float64 `json:"duration,omitempty"`
}
