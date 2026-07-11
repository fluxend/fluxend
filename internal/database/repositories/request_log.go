package repositories

import (
	"fluxend/internal/domain/logging"
	"fluxend/internal/domain/shared"
	"fmt"
	"github.com/google/uuid"
	"github.com/samber/do"
	"strings"
)

type RequestLogRepository struct {
	db shared.DB
}

func NewRequestLogRepository(injector *do.Injector) (logging.Repository, error) {
	db := do.MustInvoke[shared.DB](injector)
	return &RequestLogRepository{db: db}, nil
}

func (r *RequestLogRepository) List(
	input *logging.ListInput,
	paginationParams shared.PaginationParams,
) ([]logging.RequestLog, shared.PaginationDetails, error) {
	whereClause, params := r.buildFilters(input)

	total, err := r.getFilteredCount(whereClause, params)
	if err != nil {
		return nil, shared.PaginationDetails{}, fmt.Errorf("failed to get total count of request logs: %w", err)
	}

	requestLogs, err := r.getFilteredLogs(whereClause, params, paginationParams)
	if err != nil {
		return nil, shared.PaginationDetails{}, fmt.Errorf("failed to get request logs: %w", err)
	}

	return requestLogs, shared.PaginationDetails{
		Total: total,
		Page:  paginationParams.Page,
		Limit: paginationParams.Limit,
	}, nil
}

func (r *RequestLogRepository) buildFilters(input *logging.ListInput) (string, map[string]interface{}) {
	var filters []string
	params := make(map[string]interface{})

	filterMappings := []struct {
		condition bool
		clause    string
		paramName string
		value     interface{}
	}{
		{input.ProjectUuid.Valid, "project_uuid = :project_uuid", "project_uuid", input.ProjectUuid.UUID.String()},
		{input.UserUuid.Valid, "user_uuid = :user_uuid", "user_uuid", input.UserUuid.UUID.String()},
		{input.Status.Valid, "status = :status", "status", input.Status},
		{input.Method.Valid, "method = :method", "method", input.Method},
		{input.Endpoint.Valid, "endpoint = :endpoint", "endpoint", input.Endpoint},
		{input.IPAddress.Valid, "ip_address = :ip_address", "ip_address", input.IPAddress},
		{!input.StartTime.IsZero(), "created_at >= :date_start", "date_start", input.StartTime},
		{!input.EndTime.IsZero(), "created_at <= :date_end", "date_end", input.EndTime},
	}

	for _, mapping := range filterMappings {
		if mapping.condition {
			filters = append(filters, mapping.clause)
			params[mapping.paramName] = mapping.value
		}
	}

	whereClause := ""
	if len(filters) > 0 {
		whereClause = "WHERE " + strings.Join(filters, " AND ")
	}

	return whereClause, params
}

func (r *RequestLogRepository) getFilteredCount(whereClause string, params map[string]interface{}) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM fluxend.api_logs %s", whereClause)

	var count int
	rows, err := r.db.NamedQuery(query, params)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&count)
	}

	return count, err
}

func (r *RequestLogRepository) getFilteredLogs(whereClause string, params map[string]interface{}, paginationParams shared.PaginationParams) ([]logging.RequestLog, error) {
	// Add pagination parameters
	offset := (paginationParams.Page - 1) * paginationParams.Limit
	params["limit"] = paginationParams.Limit
	params["offset"] = offset

	sortColumn := r.validateSortColumn(paginationParams.Sort)

	// Build final query
	query := fmt.Sprintf(
		"SELECT * FROM fluxend.api_logs %s ORDER BY %s DESC LIMIT :limit OFFSET :offset",
		whereClause,
		sortColumn,
	)

	var requestLogs []logging.RequestLog
	err := r.db.SelectNamedList(&requestLogs, query, params)
	return requestLogs, err
}

func (r *RequestLogRepository) validateSortColumn(sort string) string {
	allowedSorts := map[string]bool{
		"created_at": true,
		"status":     true,
		"method":     true,
	}

	if allowedSorts[sort] {
		return sort
	}
	return "created_at" // default
}

func (r *RequestLogRepository) Create(requestLog *logging.RequestLog) (*logging.RequestLog, error) {
	return requestLog, r.db.WithTransaction(func(tx shared.Tx) error {
		query := `
        INSERT INTO fluxend.api_logs (
            project_uuid, user_uuid, api_key, method, status, endpoint, ip_address, user_agent, params, body
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        )
        RETURNING uuid
        `

		return tx.QueryRowx(
			query,
			requestLog.ProjectUuid,
			requestLog.UserUuid,
			requestLog.APIKey,
			requestLog.Method,
			requestLog.Status,
			requestLog.Endpoint,
			requestLog.IPAddress,
			requestLog.UserAgent,
			requestLog.Params,
			requestLog.Body,
		).Scan(&requestLog.Uuid)
	})
}

func (r *RequestLogRepository) CreatePostgrest(storeInput *logging.StoreInput) error {
	query := `
		INSERT INTO fluxend.api_logs (
			project_uuid, user_uuid, api_key, method, endpoint, ip_address, status, user_agent, params, body
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err := r.db.Exec(
		query,
		storeInput.ProjectUUID,
		nil,      // user_uuid
		uuid.Nil, // api_key
		storeInput.Method,
		storeInput.Endpoint,
		storeInput.IPAddress,
		200,
		"postgrest",
		"", // params
		"", // body
	)

	return err
}
