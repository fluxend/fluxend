package shared

type PostgrestService interface {
	StartContainer(dbName, jwtSecret string)
	RemoveContainer(dbName string)
	HasContainer(dbName string) bool
	RefreshSchemaCache(dbName string)
}
