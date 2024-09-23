package ldb

type DatabaseAdapter interface {
	Close() error
	Begin() (DatabaseTransaction, error)
}

type DatabaseTransaction interface {
	// perform commit; implementation may be omitted for NoSQL datbases
	Commit() error
	// perform rollback; implementation may be omitted for NoSQL databases
	Rollback() error

	SaveCollection(collection Collection) error
	DropCollection(collection Collection) error

	SaveView(view View) error
	DropView(view View) error

	// checks if the migration with the given name has already been performed
	MigrationExists(migrationName string) (bool, error)
	// saves the given migration name to the migration history
	FinishMigration(migrationName string) error

	// GetCollection(name string, fields map[string]FieldType) ([]any, error)
	// GetRecord(collection string, fields map[string]FieldType, id string) (any, error)
	// CreateRecord(collection string, fields map[string]FieldType, data map[string]any) (string, error)
	// UpdateRecord(collection string, fields map[string]FieldType, id string, data map[string]any) error
	// DeleteRecord(collection string, fields map[string]FieldType, id string) error
}
