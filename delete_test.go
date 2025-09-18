package structsql_test

import (
	"reflect"
	"testing"

	"github.com/cdvelop/structsql"
)

func TestDelete(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "DELETE FROM users WHERE id=$1"
	wantArgs := []any{1}

	s := structsql.New() // Default PostgreSQL
	var gotSQL string
	gotArgs := make([]any, 0, 10)

	err := s.Delete(u, &gotSQL, &gotArgs)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	if gotSQL != wantSQL {
		t.Fatalf("Delete SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("Delete args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
	}
}

func TestDeleteSQLite(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "DELETE FROM users WHERE id=?"
	wantArgs := []any{1}

	s := structsql.New(structsql.SQLite)
	var gotSQL string
	gotArgs := make([]any, 0, 10)

	err := s.Delete(u, &gotSQL, &gotArgs)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	if gotSQL != wantSQL {
		t.Fatalf("Delete SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("Delete args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
	}
}

func BenchmarkDelete(b *testing.B) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	s := structsql.New()
	var sql string
	args := make([]any, 0, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args = args[:0]
		_ = s.Delete(u, &sql, &args)
	}
}

func BenchmarkDeleteWithArgs(b *testing.B) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	s := structsql.New()
	var sql string
	args := make([]any, 0, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args = args[:0]
		_ = s.Delete(u, &sql, &args)
	}
}
