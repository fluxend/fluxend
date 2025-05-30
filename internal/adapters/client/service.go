package client

import (
	"fluxend/internal/database/repositories"
	"fluxend/internal/domain/database"
	"fluxend/internal/domain/shared"
	"github.com/jmoiron/sqlx"
	"github.com/samber/do"
)

type ServiceImpl struct {
	databaseRepo shared.DatabaseService
}

func NewClientService(injector *do.Injector) (database.ConnectionService, error) {
	databaseRepo := do.MustInvoke[shared.DatabaseService](injector)

	return &ServiceImpl{
		databaseRepo: databaseRepo,
	}, nil
}

func (s *ServiceImpl) ConnectByDatabaseName(name string) (*sqlx.DB, error) {
	return s.databaseRepo.Connect(name)
}

func (s *ServiceImpl) GetDatabaseStatsRepo(databaseName string, connection *sqlx.DB) (interface{}, *sqlx.DB, error) {
	clientDatabaseConnection, err := s.getOrCreateConnection(databaseName, connection)
	if err != nil {
		return nil, nil, err
	}

	clientDatabaseStatsRepo, err := repositories.NewDatabaseStatsRepository(clientDatabaseConnection)
	if err != nil {
		return nil, nil, err
	}

	return clientDatabaseStatsRepo, clientDatabaseConnection, nil
}

func (s *ServiceImpl) GetTableRepo(databaseName string, connection *sqlx.DB) (interface{}, *sqlx.DB, error) {
	clientDatabaseConnection, err := s.getOrCreateConnection(databaseName, connection)
	if err != nil {
		return nil, nil, err
	}

	clientTableRepo, err := repositories.NewTableRepository(clientDatabaseConnection)
	if err != nil {
		return nil, nil, err
	}

	return clientTableRepo, clientDatabaseConnection, nil
}

func (s *ServiceImpl) GetFunctionRepo(databaseName string, connection *sqlx.DB) (interface{}, *sqlx.DB, error) {
	clientDatabaseConnection, err := s.getOrCreateConnection(databaseName, connection)
	if err != nil {
		return nil, nil, err
	}

	clientFunctionRepo, err := repositories.NewFunctionRepository(clientDatabaseConnection)
	if err != nil {
		return nil, nil, err
	}

	return clientFunctionRepo, clientDatabaseConnection, nil
}

func (s *ServiceImpl) GetColumnRepo(databaseName string, connection *sqlx.DB) (interface{}, *sqlx.DB, error) {
	clientDatabaseConnection, err := s.getOrCreateConnection(databaseName, connection)
	if err != nil {
		return nil, nil, err
	}

	clientColumnRepo, err := repositories.NewColumnRepository(clientDatabaseConnection)
	if err != nil {
		return nil, nil, err
	}

	return clientColumnRepo, clientDatabaseConnection, nil
}

func (s *ServiceImpl) GetIndexRepo(databaseName string, connection *sqlx.DB) (interface{}, *sqlx.DB, error) {
	clientDatabaseConnection, err := s.getOrCreateConnection(databaseName, connection)
	if err != nil {
		return nil, nil, err
	}

	clientIndexRepo, err := repositories.NewIndexRepository(clientDatabaseConnection)
	if err != nil {
		return nil, nil, err
	}

	return clientIndexRepo, clientDatabaseConnection, nil
}

func (s *ServiceImpl) GetRowRepo(databaseName string, connection *sqlx.DB) (interface{}, *sqlx.DB, error) {
	clientDatabaseConnection, err := s.getOrCreateConnection(databaseName, connection)
	if err != nil {
		return nil, nil, err
	}

	clientIndexRepo, err := repositories.NewRowRepository(clientDatabaseConnection)
	if err != nil {
		return nil, nil, err
	}

	return clientIndexRepo, clientDatabaseConnection, nil
}

func (s *ServiceImpl) getOrCreateConnection(databaseName string, connection *sqlx.DB) (*sqlx.DB, error) {
	if connection != nil {
		return connection, nil
	}

	return s.databaseRepo.Connect(databaseName)
}
