package openapi

import (
	"encoding/json"
	"errors"
	"fluxend/internal/domain/auth"
	"fluxend/internal/domain/database"
	"fluxend/internal/domain/project"
	"fluxend/pkg"
	flxErrs "fluxend/pkg/errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/samber/do"
	"os"
	"strings"
)

const (
	openAPIVersion       = "3.0.0"
	apiVersion           = "1.0.0"
	apiTitleScheme       = "REST API for %s"
	apiDescriptionScheme = "Automatically generated REST API for the %s project"
	apiServerScheme      = "%s/rest/%s"
	serverTitle          = "Default API Server"
)

type Service interface {
	Generate(projectUUID uuid.UUID, requestedTables string, authUser auth.User) (string, error)
}

type ServiceImpl struct {
	projectPolicy     *project.Policy
	projectRepo       project.Repository
	connectionService database.ConnectionService
}

func NewOpenApiService(injector *do.Injector) (Service, error) {
	projectPolicy := do.MustInvoke[*project.Policy](injector)
	projectRepo := do.MustInvoke[project.Repository](injector)
	connectionService := do.MustInvoke[database.ConnectionService](injector)

	return &ServiceImpl{
		projectPolicy:     projectPolicy,
		projectRepo:       projectRepo,
		connectionService: connectionService,
	}, nil
}

func (s *ServiceImpl) Generate(projectUUID uuid.UUID, requestedTables string, authUser auth.User) (string, error) {
	fetchedProject, err := s.validateAndGetProject(projectUUID, authUser)
	if err != nil {
		return "", err
	}

	connection, clientTableRepo, clientColumnRepo, err := s.getRepositories(fetchedProject.DBName)
	if err != nil {
		return "", err
	}
	defer connection.Close()

	tables, err := clientTableRepo.List()
	if err != nil {
		return "", err
	}

	tablesToProcess := s.filterTables(tables, requestedTables)
	spec := s.generateOpenAPISpec(fetchedProject, tablesToProcess, clientColumnRepo)

	jsonBytes, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (s *ServiceImpl) validateAndGetProject(projectUUID uuid.UUID, authUser auth.User) (*project.Project, error) {
	fetchedProject, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, err
	}

	if !s.projectPolicy.CanAccess(fetchedProject.OrganizationUuid, authUser) {
		return nil, flxErrs.NewForbiddenError("project.error.viewForbidden")
	}

	return &fetchedProject, nil
}

func (s *ServiceImpl) getRepositories(dbName string) (*sqlx.DB, database.TableRepository, database.ColumnRepository, error) {
	repo, connection, err := s.connectionService.GetTableRepo(dbName, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	clientTableRepo, ok := repo.(database.TableRepository)
	if !ok {
		return connection, nil, nil, errors.New("clientTableRepo is not of type *repositories.TableRepository")
	}

	columnRepo, connection, err := s.connectionService.GetColumnRepo(dbName, connection)
	if err != nil {
		return connection, nil, nil, err
	}

	clientColumnRepo, ok := columnRepo.(database.ColumnRepository)
	if !ok {
		return connection, nil, nil, errors.New("clientColumnRepo is not of type *repositories.ColumnRepository")
	}

	return connection, clientTableRepo, clientColumnRepo, nil
}

func (s *ServiceImpl) filterTables(tables []database.Table, requestedTables string) []database.Table {
	requestedTables = strings.ReplaceAll(requestedTables, " ", "")
	if requestedTables == "" {
		return tables
	}

	requestedTableNames := strings.Split(requestedTables, ",")
	requestedSet := make(map[string]bool)
	for _, name := range requestedTableNames {
		requestedSet[name] = true
	}

	var tablesToProcess []database.Table
	for _, table := range tables {
		if requestedSet[table.Name] {
			tablesToProcess = append(tablesToProcess, table)
		}
	}

	return tablesToProcess
}

func (s *ServiceImpl) generateOpenAPISpec(project *project.Project, tables []database.Table, columnRepo database.ColumnRepository) ApiSpec {
	spec := ApiSpec{
		OpenAPI: openAPIVersion,
		Info: Info{
			Title:       fmt.Sprintf(apiTitleScheme, project.Name),
			Description: fmt.Sprintf(apiDescriptionScheme, project.Name),
			Version:     apiVersion,
		},
		Servers: []Server{
			{
				URL:         fmt.Sprintf(apiServerScheme, os.Getenv("API_URL"), project.DBName),
				Description: serverTitle,
			},
		},
		Paths: make(map[string]PathItem),
		Components: Components{
			Schemas: make(map[string]Schema),
		},
	}

	for _, table := range tables {
		columns, err := columnRepo.List(table.Name)
		if err != nil {
			continue // Skip tables with errors
		}

		s.addTableToSpec(&spec, table.Name, columns)
	}

	return spec
}

func (s *ServiceImpl) addTableToSpec(spec *ApiSpec, tableName string, columns []database.Column) {
	schema := s.generateTableSchema(columns)
	spec.Components.Schemas[tableName] = schema
	s.generateTablePaths(spec, tableName, columns)
}

func (s *ServiceImpl) generateTableSchema(columns []database.Column) Schema {
	properties := s.generateSchemaProperties(columns)
	required := s.extractRequiredFields(columns)

	return Schema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}

func (s *ServiceImpl) generateSchemaProperties(columns []database.Column) map[string]Schema {
	properties := make(map[string]Schema)

	for _, col := range columns {
		schema := s.columnToSchema(col)
		properties[col.Name] = schema
	}

	return properties
}

func (s *ServiceImpl) extractRequiredFields(columns []database.Column) []string {
	var required []string

	for _, col := range columns {
		if col.NotNull && !col.Primary {
			required = append(required, col.Name)
		}
	}

	return required
}

func (s *ServiceImpl) columnToSchema(col database.Column) Schema {
	schema := Schema{}

	switch {
	case s.isIntegerType(col.Type):
		schema.Type = "integer"
		schema.Format = s.getIntegerFormat(col.Type)
	case s.isNumericType(col.Type):
		schema.Type = "number"
	case s.isBooleanType(col.Type):
		schema.Type = "boolean"
	case s.isDateTimeType(col.Type):
		schema.Type = "string"
		schema.Format = s.getDateTimeFormat(col.Type)
	case s.isUUIDType(col.Type):
		schema.Type = "string"
		schema.Format = "uuid"
	case s.isJSONType(col.Type):
		schema.Type = "object"
	default:
		schema.Type = "string"
	}

	return schema
}

func (s *ServiceImpl) isIntegerType(colType string) bool {
	integerTypes := []string{"integer", "serial", "bigint", "bigserial", "smallint", "smallserial"}
	for _, t := range integerTypes {
		if strings.Contains(colType, t) {
			return true
		}
	}
	return false
}

func (s *ServiceImpl) getIntegerFormat(colType string) string {
	if strings.Contains(colType, "bigint") || strings.Contains(colType, "bigserial") {
		return "int64"
	}
	if strings.Contains(colType, "smallint") || strings.Contains(colType, "smallserial") {
		return "int32"
	}
	return ""
}

func (s *ServiceImpl) isNumericType(colType string) bool {
	numericTypes := []string{"decimal", "numeric", "real", "double"}
	for _, t := range numericTypes {
		if strings.Contains(colType, t) {
			return true
		}
	}
	return false
}

func (s *ServiceImpl) isBooleanType(colType string) bool {
	return strings.Contains(colType, "boolean")
}

func (s *ServiceImpl) isDateTimeType(colType string) bool {
	dateTimeTypes := []string{"timestamp", "date", "time"}
	for _, t := range dateTimeTypes {
		if strings.Contains(colType, t) {
			return true
		}
	}
	return false
}

func (s *ServiceImpl) getDateTimeFormat(colType string) string {
	if strings.Contains(colType, "timestamp") {
		return "date-time"
	}
	if strings.Contains(colType, "date") {
		return "date"
	}
	if strings.Contains(colType, "time") {
		return "time"
	}
	return ""
}

func (s *ServiceImpl) isUUIDType(colType string) bool {
	return strings.Contains(colType, "uuid")
}

func (s *ServiceImpl) isJSONType(colType string) bool {
	return strings.Contains(colType, "json")
}

func (s *ServiceImpl) generateTablePaths(spec *ApiSpec, tableName string, columns []database.Column) {
	path := "/" + tableName

	primaryKeys := s.extractPrimaryKeys(columns)

	s.addCollectionPaths(spec, path, tableName, columns)

	if len(primaryKeys) > 0 {
		s.addSingleItemPaths(spec, path, tableName, columns, primaryKeys)
	}
}

func (s *ServiceImpl) extractPrimaryKeys(columns []database.Column) []string {
	var primaryKeys []string
	for _, col := range columns {
		if col.Primary {
			primaryKeys = append(primaryKeys, col.Name)
		}
	}
	return primaryKeys
}

func (s *ServiceImpl) addCollectionPaths(spec *ApiSpec, path, tableName string, columns []database.Column) {
	getOperation := s.createGetCollectionOperation(tableName, columns)
	postOperation := s.createPostCollectionOperation(tableName)

	spec.Paths[path] = PathItem{
		Get:  getOperation,
		Post: postOperation,
	}
}

func (s *ServiceImpl) createGetCollectionOperation(tableName string, columns []database.Column) *Operation {
	operation := &Operation{
		Summary:     "Get " + tableName,
		Description: "Retrieve records from " + tableName,
		OperationID: "get" + pkg.TitleCase(tableName),
		Tags:        []string{tableName},
		Parameters:  s.createBaseParameters(),
		Responses:   s.createCollectionResponses(tableName),
	}

	operation.Parameters = append(operation.Parameters, s.createColumnFilterParameters(columns)...)
	return operation
}

func (s *ServiceImpl) createBaseParameters() []Parameter {
	return []Parameter{
		{
			Name:        "select",
			In:          "query",
			Description: "Columns to select",
			Schema:      Schema{Type: "string"},
		},
		{
			Name:        "order",
			In:          "query",
			Description: "Ordering",
			Schema:      Schema{Type: "string"},
		},
		{
			Name:        "limit",
			In:          "query",
			Description: "Limit number of results",
			Schema:      Schema{Type: "integer"},
		},
		{
			Name:        "offset",
			In:          "query",
			Description: "Offset for pagination",
			Schema:      Schema{Type: "integer"},
		},
	}
}

func (s *ServiceImpl) createColumnFilterParameters(columns []database.Column) []Parameter {
	var parameters []Parameter
	for _, col := range columns {
		parameters = append(parameters, Parameter{
			Name:        col.Name,
			In:          "query",
			Description: "Filter by " + col.Name,
			Schema:      s.columnToSchema(col),
		})
	}
	return parameters
}

func (s *ServiceImpl) createCollectionResponses(tableName string) map[string]Response {
	return map[string]Response{
		"200": {
			Description: "Success",
			Content: map[string]MediaType{
				"application/json": {
					Schema: Schema{
						Type: "array",
						Items: &Schema{
							Ref: "#/components/schemas/" + tableName,
						},
					},
				},
			},
		},
	}
}

func (s *ServiceImpl) createPostCollectionOperation(tableName string) *Operation {
	return &Operation{
		Summary:     "Create " + tableName,
		Description: "Create new record in " + tableName,
		OperationID: "create" + pkg.TitleCase(tableName),
		Tags:        []string{tableName},
		RequestBody: s.createRequestBody(tableName),
		Responses:   s.createSingleItemResponses(tableName, "201", "Created"),
	}
}

func (s *ServiceImpl) createRequestBody(tableName string) *RequestBody {
	return &RequestBody{
		Description: "Record to create",
		Required:    true,
		Content: map[string]MediaType{
			"application/json": {
				Schema: Schema{
					Ref: "#/components/schemas/" + tableName,
				},
			},
		},
	}
}

func (s *ServiceImpl) createSingleItemResponses(tableName, statusCode, description string) map[string]Response {
	return map[string]Response{
		statusCode: {
			Description: description,
			Content: map[string]MediaType{
				"application/json": {
					Schema: Schema{
						Ref: "#/components/schemas/" + tableName,
					},
				},
			},
		},
	}
}

func (s *ServiceImpl) addSingleItemPaths(spec *ApiSpec, path, tableName string, columns []database.Column, primaryKeys []string) {
	singlePath := s.buildSingleItemPath(path, primaryKeys)
	parameters := s.createPrimaryKeyParameters(primaryKeys, columns)

	getSingle := s.createGetSingleOperation(tableName, parameters)
	patchSingle := s.createPatchSingleOperation(tableName, parameters)
	deleteSingle := s.createDeleteSingleOperation(tableName, parameters)

	spec.Paths[singlePath] = PathItem{
		Get:    getSingle,
		Patch:  patchSingle,
		Delete: deleteSingle,
	}
}

func (s *ServiceImpl) buildSingleItemPath(path string, primaryKeys []string) string {
	var pathParams []string
	for _, pk := range primaryKeys {
		pathParams = append(pathParams, "{"+pk+"}")
	}
	return path + "?" + strings.Join(pathParams, "&")
}

func (s *ServiceImpl) createPrimaryKeyParameters(primaryKeys []string, columns []database.Column) []Parameter {
	var parameters []Parameter

	for _, pk := range primaryKeys {
		pkSchema := s.findColumnSchema(pk, columns)
		parameters = append(parameters, Parameter{
			Name:     pk,
			In:       "path",
			Required: true,
			Schema:   pkSchema,
		})
	}

	return parameters
}

func (s *ServiceImpl) findColumnSchema(columnName string, columns []database.Column) Schema {
	for _, col := range columns {
		if col.Name == columnName {
			return s.columnToSchema(col)
		}
	}
	return Schema{Type: "string"} // fallback
}

func (s *ServiceImpl) createGetSingleOperation(tableName string, parameters []Parameter) *Operation {
	return &Operation{
		Summary:     "Get single " + tableName,
		Description: "Retrieve a single record from " + tableName,
		OperationID: "get" + pkg.TitleCase(tableName) + "ById",
		Tags:        []string{tableName},
		Parameters:  parameters,
		Responses:   s.createSingleItemResponses(tableName, "200", "Success"),
	}
}

func (s *ServiceImpl) createPatchSingleOperation(tableName string, parameters []Parameter) *Operation {
	return &Operation{
		Summary:     "Update " + tableName,
		Description: "Update a record in " + tableName,
		OperationID: "update" + pkg.TitleCase(tableName),
		Tags:        []string{tableName},
		Parameters:  parameters,
		RequestBody: &RequestBody{
			Description: "Fields to update",
			Content: map[string]MediaType{
				"application/json": {
					Schema: Schema{
						Ref: "#/components/schemas/" + tableName,
					},
				},
			},
		},
		Responses: s.createSingleItemResponses(tableName, "200", "Updated"),
	}
}

func (s *ServiceImpl) createDeleteSingleOperation(tableName string, parameters []Parameter) *Operation {
	return &Operation{
		Summary:     "Delete " + tableName,
		Description: "Delete a record from " + tableName,
		OperationID: "delete" + pkg.TitleCase(tableName),
		Tags:        []string{tableName},
		Parameters:  parameters,
		Responses: map[string]Response{
			"204": {
				Description: "Deleted",
			},
		},
	}
}
