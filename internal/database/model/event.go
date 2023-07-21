package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Event struct {
	ID         uuid.UUID         `gorm:"default:uuid_generate_v4()" json:"event_id"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"-"`
	DeletedAt  gorm.DeletedAt    `json:"-"`
	StreamName string            `json:"-"`
	Type       string            `json:"type"`
	Data       datatypes.JSONMap `json:"data" gorm:"type:jsonb"`
}

type TransferEvent struct {
	Type string                 `json:"event_type"`
	Data map[string]interface{} `json:"event_data"`
}
