package shared

type PostgrestService interface {
	Start(dbName, jwtSecret string)
	StartAll()
	Stop(dbName string)
	IsRunning(dbName string) bool
	Endpoint(dbName string) (string, bool)
	RefreshSchemaCache(dbName string)
}
