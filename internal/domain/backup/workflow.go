package backup

import (
	"fluxend/internal/adapters/storage"
	"fluxend/internal/config/constants"
	"fluxend/internal/domain/setting"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"os"
	"os/exec"
	"time"
)

type WorkflowService interface {
	Create(databaseName string, backupUUID uuid.UUID)
	Delete(databaseName string, backupUUID uuid.UUID)
}

type WorkflowServiceImpl struct {
	settingService setting.Service
	backupRepo     Repository
	storageFactory *storage.Factory
}

func NewBackupWorkflowService(injector *do.Injector) (WorkflowService, error) {
	settingService, err := setting.NewSettingService(injector)
	if err != nil {
		return nil, err
	}

	backupRepo := do.MustInvoke[Repository](injector)
	storageFactory := do.MustInvoke[*storage.Factory](injector)

	return &WorkflowServiceImpl{
		settingService: settingService,
		backupRepo:     backupRepo,
		storageFactory: storageFactory,
	}, nil
}

// Create pg_dump, copy file, ensure container exists, and upload to S3
func (s *WorkflowServiceImpl) Create(databaseName string, backupUUID uuid.UUID) {
	backupFilePath := fmt.Sprintf("/tmp/%s.sql", backupUUID)

	// 1. Execute pg_dump
	if err := s.executePgDump(databaseName, backupFilePath); err != nil {
		s.handleBackupFailure(backupUUID, constants.BackupStatusCreatingFailed, err.Error())

		return
	}

	// 2. Ensure backup container exists
	if err := s.ensureBackupContainerExists(); err != nil {
		s.handleBackupFailure(backupUUID, constants.BackupStatusCreatingFailed, err.Error())

		return
	}

	fileBytes, err := s.readBackupFile(backupUUID)
	if err != nil {
		s.handleBackupFailure(backupUUID, constants.BackupStatusCreatingFailed, err.Error())

		return
	}

	// 3. Upload backup to S3
	if err = s.uploadBackup(databaseName, backupUUID, fileBytes); err != nil {
		s.handleBackupFailure(backupUUID, constants.BackupStatusCreatingFailed, err.Error())

		return
	}

	// 4. Update backup status to completed
	err = s.backupRepo.UpdateStatus(backupUUID, constants.BackupStatusCreated, "", time.Now())
	if err != nil {
		s.handleBackupFailure(backupUUID, constants.BackupStatusCreatingFailed, err.Error())
	}

	// 5. Remove backup file from app container
	if err = os.Remove(backupFilePath); err != nil {
		log.Error().
			Str("action", constants.ActionBackup).
			Str("db", databaseName).
			Str("backup_uuid", backupUUID.String()).
			Str("error", err.Error()).
			Msg("failed to remove backup file from fluxend_api container")
	}
}

func (s *WorkflowServiceImpl) Delete(databaseName string, backupUUID uuid.UUID) {
	filePath := fmt.Sprintf("%s/%s.sql", databaseName, backupUUID)

	storageService, err := s.storageFactory.CreateProvider(s.settingService.GetStorageDriver())
	if err != nil {
		log.Error().
			Str("action", constants.ActionBackup).
			Str("db", databaseName).
			Str("backup_uuid", backupUUID.String()).
			Str("error", err.Error()).
			Msg("failed to get storage provider")
		s.handleBackupFailure(backupUUID, constants.BackupStatusDeletingFailed, err.Error())

		return
	}

	if err = storageService.DeleteFile(storage.FileInput{
		ContainerName: constants.BackupContainerName,
		FileName:      filePath,
	}); err != nil {
		s.handleBackupFailure(backupUUID, constants.BackupStatusDeletingFailed, err.Error())
	}

	_, err = s.backupRepo.Delete(backupUUID)
	if err != nil {
		log.Error().
			Str("action", constants.ActionBackup).
			Str("db", databaseName).
			Str("backup_uuid", backupUUID.String()).
			Str("error", err.Error()).
			Msg("failed to delete backup from database")
	}
}

func (s *WorkflowServiceImpl) executePgDump(databaseName, backupFilePath string) error {
	cmd := exec.Command(
		"pg_dump",
		"-h", os.Getenv("DATABASE_HOST"),
		"-U", os.Getenv("DATABASE_USER"),
		"-d", databaseName,
		"-f", backupFilePath,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+os.Getenv("DATABASE_PASSWORD"))

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Error().
			Str("action", constants.ActionBackup).
			Str("db", databaseName).
			Str("output", string(output)).
			Str("error", err.Error()).
			Msg("failed to execute pg_dump command")

		return err
	}

	return nil
}

func (s *WorkflowServiceImpl) ensureBackupContainerExists() error {
	storageService, err := s.storageFactory.CreateProvider(s.settingService.GetStorageDriver())
	if err != nil {
		log.Error().
			Str("action", constants.ActionBackup).
			Str("error", err.Error()).
			Msg("failed to get storage provider")

		return err
	}

	containerExists := storageService.ContainerExists(constants.BackupContainerName)
	if !containerExists {
		_, err := storageService.CreateContainer(constants.BackupContainerName)
		if err != nil {
			log.Error().
				Str("action", constants.ActionBackup).
				Str("container_name", constants.BackupContainerName).
				Str("error", err.Error()).
				Msg("failed to create backup container")

			return err
		}
	}

	return nil
}

func (s *WorkflowServiceImpl) readBackupFile(backupUUID uuid.UUID) ([]byte, error) {
	fileBytes, err := os.ReadFile(fmt.Sprintf("/tmp/%s.sql", backupUUID))
	if err != nil {
		log.Error().
			Str("action", constants.ActionBackup).
			Str("backup_uuid", backupUUID.String()).
			Str("error", err.Error()).
			Msg("failed to read backup file from fluxend_api container")

		return nil, err
	}

	return fileBytes, err
}

func (s *WorkflowServiceImpl) uploadBackup(databaseName string, backupUUID uuid.UUID, fileBytes []byte) error {
	filePath := fmt.Sprintf("%s/%s.sql", databaseName, backupUUID)

	storageService, err := s.storageFactory.CreateProvider(s.settingService.GetStorageDriver())
	if err != nil {
		log.Error().
			Str("action", constants.ActionBackup).
			Str("db", databaseName).
			Str("backup_uuid", backupUUID.String()).
			Str("error", err.Error()).
			Msg("failed to get storage provider")

		return err
	}

	err = storageService.UploadFile(storage.UploadFileInput{
		ContainerName: constants.BackupContainerName,
		FileName:      filePath,
		FileBytes:     fileBytes,
	})
	if err != nil {
		log.Error().
			Str("action", constants.ActionBackup).
			Str("db", databaseName).
			Str("backup_uuid", backupUUID.String()).
			Str("error", err.Error()).
			Msg("failed to upload backup file to S3")

		return err
	}

	return err
}

// handleBackupFailure updates the backup status to appropriate state and logs the error
func (s *WorkflowServiceImpl) handleBackupFailure(backupUUID uuid.UUID, status, errorMessage string) {
	err := s.backupRepo.UpdateStatus(backupUUID, status, errorMessage, time.Now())
	if err != nil {
		log.Error().
			Str("action", constants.ActionBackup).
			Str("backup_uuid", backupUUID.String()).
			Str("error", err.Error()).
			Msg("failed to update backup status in database")
		return
	}

	log.Error().
		Str("action", constants.ActionBackup).
		Str("backup_uuid", backupUUID.String()).
		Str("status", status).
		Str("error", errorMessage).
		Msg("backup workflow failed")
}
