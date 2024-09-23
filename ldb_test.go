package ldb_test

import (
	"testing"

	"lehnert.dev/ldb"
)

func TestSQLite(t *testing.T) {
	adapter, err := ldb.OpenDuckDBAdapter("/tmp/test.db")
	if err != nil {
		t.Fatal(err)
	}

	tx, err := adapter.Begin()
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.SaveCollection(ldb.Collection{
		Name: "test0",
		Schema: &ldb.CollectionSchema{
			Fields: []*ldb.Field{
				{
					Name: "id",
					Schema: &ldb.FieldSchema{
						Type: ldb.FieldTypeId{
							PrimaryKey: true,
						},
					},
				},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := tx.SaveCollection(ldb.Collection{
		Name: "test1",
		Schema: &ldb.CollectionSchema{
			Fields: []*ldb.Field{
				{Name: "bool", Schema: &ldb.FieldSchema{Type: ldb.FieldTypeBool{}}},
				{Name: "datetime", Schema: &ldb.FieldSchema{Type: ldb.FieldTypeDateTime{}}},
				{Name: "enum", Schema: &ldb.FieldSchema{Type: ldb.FieldTypeEnum{EnumValues: []string{"a", "b", "c"}}}},
				{Name: "float", Schema: &ldb.FieldSchema{Type: ldb.FieldTypeFloat{}}},
				{Name: "id", Schema: &ldb.FieldSchema{Type: ldb.FieldTypeId{PrimaryKey: true}}},
				{Name: "int", Schema: &ldb.FieldSchema{Type: ldb.FieldTypeInt{}}},
				{Name: "singleRelation", Schema: &ldb.FieldSchema{Type: ldb.FieldTypeSingleRelation{Collection: "test0"}}},
				{Name: "text", Schema: &ldb.FieldSchema{Type: ldb.FieldTypeText{}}},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	if err := adapter.Close(); err != nil {
		t.Fatal(err)
	}
}
