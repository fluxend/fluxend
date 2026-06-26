package repositories

import (
	"fluxend/internal/domain/shared"
	"fluxend/internal/domain/webhook"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/samber/do"
)

type WebhookRepository struct {
	db shared.DB
}

func NewWebhookRepository(injector *do.Injector) (webhook.Repository, error) {
	db := do.MustInvoke[shared.DB](injector)
	return &WebhookRepository{db: db}, nil
}

func (r *WebhookRepository) ListForTable(projectUUID uuid.UUID, tableName string) ([]webhook.Config, error) {
	query := `
		SELECT uuid, project_uuid, table_name, url, events, is_active, created_by, updated_by, created_at, updated_at
		FROM fluxend.webhook_configs
		WHERE project_uuid = $1 AND table_name = $2
		ORDER BY created_at DESC
	`

	var configs []webhook.Config
	if err := r.db.Select(&configs, query, projectUUID, tableName); err != nil {
		return nil, err
	}

	return configs, nil
}

func (r *WebhookRepository) ListActiveForTableAndEvent(projectUUID uuid.UUID, tableName, event string) ([]webhook.Config, error) {
	query := `
		SELECT uuid, project_uuid, table_name, url, events, is_active, created_by, updated_by, created_at, updated_at
		FROM fluxend.webhook_configs
		WHERE project_uuid = $1 AND table_name = $2 AND is_active = TRUE AND $3 = ANY(events)
	`

	var configs []webhook.Config
	if err := r.db.Select(&configs, query, projectUUID, tableName, event); err != nil {
		return nil, err
	}

	return configs, nil
}

func (r *WebhookRepository) GetByUUID(webhookUUID uuid.UUID) (webhook.Config, error) {
	query := `
		SELECT uuid, project_uuid, table_name, url, events, is_active, created_by, updated_by, created_at, updated_at
		FROM fluxend.webhook_configs
		WHERE uuid = $1
	`

	var config webhook.Config
	return config, r.db.GetWithNotFound(&config, "webhook.error.notFound", query, webhookUUID)
}

func (r *WebhookRepository) ExistsByUUID(webhookUUID uuid.UUID) (bool, error) {
	return r.db.Exists("fluxend.webhook_configs", "uuid = $1", webhookUUID)
}

func (r *WebhookRepository) Create(config *webhook.Config) (*webhook.Config, error) {
	return config, r.db.WithTransaction(func(tx shared.Tx) error {
		query := `
			INSERT INTO fluxend.webhook_configs (
				project_uuid, table_name, url, events, is_active, created_by, updated_by
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7
			)
			RETURNING uuid
		`

		return tx.QueryRowx(
			query,
			config.ProjectUuid,
			config.TableName,
			config.URL,
			pq.Array(config.Events),
			config.IsActive,
			config.CreatedBy,
			config.UpdatedBy,
		).Scan(&config.Uuid)
	})
}

func (r *WebhookRepository) Delete(webhookUUID uuid.UUID) (bool, error) {
	rowsAffected, err := r.db.ExecWithRowsAffected(
		"DELETE FROM fluxend.webhook_configs WHERE uuid = $1", webhookUUID,
	)
	if err != nil {
		return false, err
	}

	return rowsAffected == 1, nil
}
