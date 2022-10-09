package model

import (
	"time"
)

type EventView struct {
	LastModified time.Time `json:"last_modified"`
	Entries      []Event   `json:"entries"`
}
