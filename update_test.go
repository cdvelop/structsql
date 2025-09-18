package structsql_test

import (
	"reflect"
	"testing"

	"github.com/cdvelop/structsql"
)

func TestUpdate(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "UPDATE users SET name=$1, email=$2 WHERE id=$3"
	wantArgs := []any{"Alice", "alice@example.com", 1}

	s := structsql.New() // Default PostgreSQL
	var gotSQL string
	gotArgs := make([]any, 0, 10)

	err := s.Update(u, &gotSQL, &gotArgs)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if gotSQL != wantSQL {
		t.Fatalf("Update SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("Update args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
	}
}

func TestUpdateSQLite(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "UPDATE users SET name=?, email=? WHERE id=?"
	wantArgs := []any{"Alice", "alice@example.com", 1}

	s := structsql.New(structsql.SQLite)
	var gotSQL string
	gotArgs := make([]any, 0, 10)

	err := s.Update(u, &gotSQL, &gotArgs)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if gotSQL != wantSQL {
		t.Fatalf("Update SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("Update args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
	}
}

func BenchmarkUpdate(b *testing.B) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	s := structsql.New()
	var sql string
	args := make([]any, 0, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args = args[:0]
		_ = s.Update(u, &sql, &args)
	}
}

func BenchmarkUpdateWithArgs(b *testing.B) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	s := structsql.New()
	var sql string
	args := make([]any, 0, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args = args[:0]
		_ = s.Update(u, &sql, &args)
	}
}
