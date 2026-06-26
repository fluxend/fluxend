package webhook

import (
	"bytes"
	"encoding/json"
	"fluxend/internal/domain/auth"
	"fluxend/internal/domain/project"
	"fluxend/pkg/errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"net/http"
	"time"
)

type Service interface {
	List(projectUUID uuid.UUID, tableName string, authUser auth.User) ([]Config, error)
	Create(input CreateInput, authUser auth.User) (Config, error)
	Delete(webhookUUID uuid.UUID, authUser auth.User) (bool, error)
	Fire(projectUUID uuid.UUID, tableName, event string, record interface{})
}

type CreateInput struct {
	ProjectUUID uuid.UUID
	TableName   string
	URL         string
	Events      []string
	IsActive    bool
}

type ServiceImpl struct {
	policy      *Policy
	webhookRepo Repository
	projectRepo project.Repository
	httpClient  *http.Client
}

func NewWebhookService(injector *do.Injector) (Service, error) {
	policy := do.MustInvoke[*Policy](injector)
	webhookRepo := do.MustInvoke[Repository](injector)
	projectRepo := do.MustInvoke[project.Repository](injector)

	return &ServiceImpl{
		policy:      policy,
		webhookRepo: webhookRepo,
		projectRepo: projectRepo,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (s *ServiceImpl) List(projectUUID uuid.UUID, tableName string, authUser auth.User) ([]Config, error) {
	organizationUUID, err := s.projectRepo.GetOrganizationUUIDByProjectUUID(projectUUID)
	if err != nil {
		return nil, err
	}

	if !s.policy.CanAccess(organizationUUID, authUser) {
		return nil, errors.NewForbiddenError("webhook.error.listForbidden")
	}

	return s.webhookRepo.ListForTable(projectUUID, tableName)
}

func (s *ServiceImpl) Create(input CreateInput, authUser auth.User) (Config, error) {
	organizationUUID, err := s.projectRepo.GetOrganizationUUIDByProjectUUID(input.ProjectUUID)
	if err != nil {
		return Config{}, err
	}

	if !s.policy.CanCreate(organizationUUID, authUser) {
		return Config{}, errors.NewForbiddenError("webhook.error.createForbidden")
	}

	config := Config{
		ProjectUuid: input.ProjectUUID,
		TableName:   input.TableName,
		URL:         input.URL,
		Events:      pq.StringArray(input.Events),
		IsActive:    input.IsActive,
		CreatedBy:   authUser.Uuid,
		UpdatedBy:   authUser.Uuid,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = s.webhookRepo.Create(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func (s *ServiceImpl) Delete(webhookUUID uuid.UUID, authUser auth.User) (bool, error) {
	fetched, err := s.webhookRepo.GetByUUID(webhookUUID)
	if err != nil {
		return false, err
	}

	organizationUUID, err := s.projectRepo.GetOrganizationUUIDByProjectUUID(fetched.ProjectUuid)
	if err != nil {
		return false, err
	}

	if !s.policy.CanDelete(organizationUUID, authUser) {
		return false, errors.NewForbiddenError("webhook.error.deleteForbidden")
	}

	return s.webhookRepo.Delete(webhookUUID)
}

func (s *ServiceImpl) Fire(projectUUID uuid.UUID, tableName, event string, record interface{}) {
	webhooks, err := s.webhookRepo.ListActiveForTableAndEvent(projectUUID, tableName, event)
	if err != nil || len(webhooks) == 0 {
		return
	}

	payload := Payload{
		Event:     event,
		Table:     tableName,
		Project:   projectUUID.String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Record:    record,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Error().Str("table", tableName).Str("event", event).Msg("webhook: failed to marshal payload")
		return
	}

	for _, wh := range webhooks {
		if err := s.post(wh.URL, body); err != nil {
			log.Error().
				Str("url", wh.URL).
				Str("table", tableName).
				Str("event", event).
				Str("error", err.Error()).
				Msg("webhook: delivery failed")
		}
	}
}

func (s *ServiceImpl) post(url string, body []byte) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
