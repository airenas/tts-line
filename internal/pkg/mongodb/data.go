package mongodb

import (
	"time"
)

//TextRecord data in mongo db
type TextRecord struct {
	ID      string    `json:"id"`
	Type    int       `json:"type"`
	Text    string    `json:"text"`
	Created time.Time `json:"created,omitempty"`
	Tags    []string  `json:"tags,omitempty"`
}
