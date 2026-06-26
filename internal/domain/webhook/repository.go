package webhook

import "github.com/google/uuid"

type Repository interface {
	ListForTable(projectUUID uuid.UUID, tableName string) ([]Config, error)
	ListActiveForTableAndEvent(projectUUID uuid.UUID, tableName, event string) ([]Config, error)
	GetByUUID(webhookUUID uuid.UUID) (Config, error)
	ExistsByUUID(webhookUUID uuid.UUID) (bool, error)
	Create(config *Config) (*Config, error)
	Delete(webhookUUID uuid.UUID) (bool, error)
}
