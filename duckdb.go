package ldb

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/samber/lo"
)

var _ DatabaseAdapter = DuckDBAdapter{}
var _ DatabaseTransaction = DuckDBTransaction{}

type DuckDBAdapter struct {
	db *sql.DB
}

func OpenDuckDBAdapter(databaseFilePath string) (*DuckDBAdapter, error) {
	db, err := sql.Open("duckdb", databaseFilePath)
	if err != nil {
		return nil, err
	}

	return &DuckDBAdapter{db}, nil
}

func (s DuckDBAdapter) Close() error {
	return s.db.Close()
}

func (s DuckDBAdapter) Begin() (DatabaseTransaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	return DatabaseTransaction(DuckDBTransaction{tx}), nil
}

type DuckDBTransaction struct {
	tx *sql.Tx
}

// Commit implements DatabaseTransaction.
func (s DuckDBTransaction) Commit() error {
	return s.tx.Commit()
}

// Rollback implements DatabaseTransaction.
func (s DuckDBTransaction) Rollback() error {
	return s.tx.Rollback()
}

// SaveCollection implements DatabaseTransaction.
func (s DuckDBTransaction) SaveCollection(collection Collection) error {
	// create collection if not exists
	if collection.original == nil {
		columns := []string{}
		for _, field := range collection.Schema.Fields {
			columns = append(columns, columnSQL(field.Name, field.Schema.Type))
		}

		sql := fmt.Sprintf("CREATE TABLE %s (%s)", collection.Name, strings.Join(columns, ", "))

		_, err := s.tx.Exec(sql)
		return err
	}

	// rename collection if neccessary
	if collection.original.Name != collection.Name {
		sql := fmt.Sprintf("ALTER TABLE %s RENAME TO %s", collection.original.Name, collection.Name)
		_, err := s.tx.Exec(sql)
		if err != nil {

			return err
		}
	}

	createFields := lo.Filter(collection.Schema.Fields, func(field *Field, i int) bool {
		return field.original == nil
	})

	renameFields := lo.Filter(collection.Schema.Fields, func(field *Field, i int) bool {
		return field.original.original.Name != field.Name
	})

	removeFields := []*Field{}
	if collection.original != nil {
		removeFields = lo.Filter(collection.original.Schema.Fields, func(origField *Field, i int) bool {
			_, found := lo.Find(collection.Schema.Fields, func(field *Field) bool {
				return field.original != nil && field.original.Name == origField.Name
			})

			return !found
		})
	}

	for _, field := range removeFields {
		sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", collection.Name, field.Name)
		if _, err := s.tx.Exec(sql); err != nil {
			return err
		}
	}

	for _, field := range renameFields {
		sql := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", collection.Name, field.original.Name, field.Name)
		if _, err := s.tx.Exec(sql); err != nil {
			return err
		}
	}

	for _, field := range createFields {
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", collection.Name, columnSQL(field.Name, field.Schema.Type))
		if _, err := s.tx.Exec(sql); err != nil {
			return err
		}
	}

	return nil
}

// DropCollection implements DatabaseTransaction.
func (s DuckDBTransaction) DropCollection(collection Collection) error {
	panic("unimplemented")
}

// SaveView implements DatabaseTransaction.
func (s DuckDBTransaction) SaveView(view View) error {
	panic("unimplemented")
}

// DropView implements DatabaseTransaction.
func (s DuckDBTransaction) DropView(view View) error {
	panic("unimplemented")
}

// MigrationExists implements DatabaseTransaction.
func (s DuckDBTransaction) MigrationExists(migrationName string) (bool, error) {
	panic("unimplemented")
}

// FinishMigration implements DatabaseTransaction.
func (s DuckDBTransaction) FinishMigration(migrationName string) error {
	panic("unimplemented")
}

func withNullConstraint(sql string, nullable bool) string {
	if nullable {
		return sql + " NULL"
	}

	return sql + " NOT NULL"
}

func columnSQL(column string, fieldType FieldType) string {
	switch ft := fieldType.(type) {
	case FieldTypeBool:
		return withNullConstraint(column+" BOOL", ft.Nullable)

	case FieldTypeDateTime:
		return withNullConstraint(column+" TIMESTAMP", ft.Nullable)

	case FieldTypeEnum:
		return withNullConstraint(column+" TEXT", ft.Nullable)

	case FieldTypeFloat:
		return withNullConstraint(column+" REAL", ft.Nullable)

	case FieldTypeId:
		sql := withNullConstraint(column+" TEXT", ft.Nullable || ft.PrimaryKey)

		if ft.PrimaryKey {
			sql += " PRIMARY KEY"
		}

		return sql

	case FieldTypeInt:
		return withNullConstraint(column+" BIGINT", ft.Nullable)

	case FieldTypeSingleRelation:
		sql := withNullConstraint(column+" TEXT", ft.Nullable)
		sql += " REFERENCES " + ft.Collection + "(id)"

		if ft.CascadeDelete {
			sql += " ON DELETE CASCADE"
		}

		return sql

	case FieldTypeText:
		return withNullConstraint(column+" TEXT", ft.Nullable)

	default:
		panic("SQLiteAdapter: unexpected fieldType")
	}
}
