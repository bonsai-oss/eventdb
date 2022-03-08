package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/datatypes"

	"github.com/google/uuid"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JSONB) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

type Event struct {
	ID         uuid.UUID `gorm:"default:uuid_generate_v4()" json:"event_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"-"`
	DeletedAt  time.Time `json:"-"`
	StreamName string    `json:"-"`
	TransferEvent
}

type TransferEvent struct {
	Type string         `json:"event_type"`
	Data datatypes.JSON `json:"event_data"`
}
