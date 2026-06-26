package app

import (
	"fluxend/internal/adapters/client"
	"fluxend/internal/adapters/email"
	"fluxend/internal/adapters/postgrest"
	sqlxAdapter "fluxend/internal/adapters/sqlx"
	"fluxend/internal/adapters/storage"
	"fluxend/internal/api/handlers"
	"fluxend/internal/database"
	"fluxend/internal/database/factories"
	"fluxend/internal/database/repositories"
	"fluxend/internal/domain/backup"
	databaseDomain "fluxend/internal/domain/database"
	"fluxend/internal/domain/form"
	"fluxend/internal/domain/health"
	"fluxend/internal/domain/logging"
	"fluxend/internal/domain/openapi"
	"fluxend/internal/domain/organization"
	"fluxend/internal/domain/project"
	"fluxend/internal/domain/setting"
	"fluxend/internal/domain/shared"
	"fluxend/internal/domain/stats"
	"fluxend/internal/domain/storage/container"
	"fluxend/internal/domain/storage/file"
	"fluxend/internal/domain/user"
	"fluxend/internal/domain/webhook"
	"github.com/jmoiron/sqlx"
	"github.com/samber/do"
)

func InitializeContainer() *do.Injector {
	injector := do.New()

	// --- Database ---
	database.InitDB()
	do.Provide(injector, func(i *do.Injector) (*sqlx.DB, error) {
		return database.GetDB(), nil
	})
	do.Provide(injector, func(i *do.Injector) (shared.DB, error) {
		return sqlxAdapter.NewAdapter(do.MustInvoke[*sqlx.DB](i)), nil
	})

	// --- Logging ---
	do.Provide(injector, logging.NewLogService)
	do.Provide(injector, repositories.NewRequestLogRepository)
	do.Provide(injector, client.NewDatabaseRepository)

	// --- User ---
	do.Provide(injector, user.NewUserPolicy)
	do.Provide(injector, repositories.NewUserRepository)
	do.Provide(injector, user.NewUserService)
	do.Provide(injector, handlers.NewUserHandler)
	do.Provide(injector, factories.NewUserFactory)

	// --- Setting ---
	do.Provide(injector, repositories.NewSettingRepository)
	do.Provide(injector, setting.NewSettingService)
	do.Provide(injector, handlers.NewSettingHandler)

	// --- Organization ---
	do.Provide(injector, organization.NewOrganizationPolicy)
	do.Provide(injector, repositories.NewOrganizationRepository)
	do.Provide(injector, organization.NewOrganizationService)
	do.Provide(injector, handlers.NewOrganizationHandler)
	do.Provide(injector, handlers.NewOrganizationMemberHandler)

	// --- Project ---
	do.Provide(injector, project.NewProjectPolicy)
	do.Provide(injector, repositories.NewProjectRepository)
	do.Provide(injector, project.NewProjectService)
	do.Provide(injector, openapi.NewOpenApiService)
	do.Provide(injector, handlers.NewProjectHandler)

	// --- Forms ---
	do.Provide(injector, repositories.NewFormRepository)
	do.Provide(injector, repositories.NewFormFieldRepository)
	do.Provide(injector, repositories.NewFormResponseRepository)

	do.Provide(injector, form.NewFormService)
	do.Provide(injector, form.NewFormFieldValidationService)
	do.Provide(injector, form.NewFieldService)
	do.Provide(injector, form.NewFormResponseService)

	do.Provide(injector, handlers.NewFormHandler)
	do.Provide(injector, handlers.NewFormFieldHandler)
	do.Provide(injector, handlers.NewFormResponseHandler)

	// --- Storage ---
	do.Provide(injector, repositories.NewContainerRepository)
	do.Provide(injector, repositories.NewFileRepository)

	do.Provide(injector, container.NewContainerService)
	do.Provide(injector, file.NewFileService)

	do.Provide(injector, handlers.NewContainerHandler)
	do.Provide(injector, handlers.NewFileHandler)

	// --- Backups ---
	do.Provide(injector, repositories.NewBackupRepository)
	do.Provide(injector, backup.NewBackupWorkflowService)
	do.Provide(injector, backup.NewBackupService)
	do.Provide(injector, handlers.NewBackupHandler)

	// --- Client & Stats ---
	do.Provide(injector, client.NewClientService)
	do.Provide(injector, stats.NewDatabaseStatsService)
	do.Provide(injector, postgrest.NewPostgrestService)
	do.Provide(injector, handlers.NewStatHandler)

	// --- Tables ---
	do.Provide(injector, databaseDomain.NewTableService)
	do.Provide(injector, databaseDomain.NewFileImportService)
	do.Provide(injector, databaseDomain.NewColumnService)
	do.Provide(injector, databaseDomain.NewIndexService)
	do.Provide(injector, databaseDomain.NewFunctionService)

	do.Provide(injector, handlers.NewTableHandler)
	do.Provide(injector, handlers.NewColumnHandler)
	do.Provide(injector, handlers.NewIndexHandler)
	do.Provide(injector, handlers.NewFunctionHandler)

	// --- Webhooks ---
	do.Provide(injector, repositories.NewWebhookRepository)
	do.Provide(injector, webhook.NewWebhookPolicy)
	do.Provide(injector, webhook.NewWebhookService)
	do.Provide(injector, handlers.NewWebhookHandler)
	do.Provide(injector, handlers.NewRowHandler)

	// --- Health ---
	do.Provide(injector, health.NewHealthService)
	do.Provide(injector, handlers.NewHealthHandler)

	do.Provide(injector, storage.NewFactory)
	do.Provide(injector, email.NewFactory)

	return injector
}
