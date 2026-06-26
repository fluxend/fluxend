package webhook

import (
	"github.com/google/uuid"
)

type Response struct {
	Uuid      uuid.UUID `json:"uuid"`
	TableName string    `json:"tableName"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	IsActive  bool      `json:"isActive"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}
