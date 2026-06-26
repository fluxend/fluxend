package webhook

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"time"
)

type Config struct {
	Uuid        uuid.UUID      `db:"uuid"`
	ProjectUuid uuid.UUID      `db:"project_uuid"`
	TableName   string         `db:"table_name"`
	URL         string         `db:"url"`
	Events      pq.StringArray `db:"events"`
	IsActive    bool           `db:"is_active"`
	CreatedBy   uuid.UUID      `db:"created_by"`
	UpdatedBy   uuid.UUID      `db:"updated_by"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

type Payload struct {
	Event     string      `json:"event"`
	Table     string      `json:"table"`
	Project   string      `json:"project"`
	Timestamp string      `json:"timestamp"`
	Record    interface{} `json:"record"`
}
