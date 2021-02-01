package mongodb

import (
	"time"
)

type textRecord struct {
	ID      string    `json:"id"`
	Type    int       `json:"type"`
	Text    string    `json:"text"`
	Created time.Time `json:"created,omitempty"`
}
