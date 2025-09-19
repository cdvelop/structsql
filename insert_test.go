package structsql_test

import (
	"reflect"
	"testing"

	"github.com/cdvelop/structsql"
)

func TestInsert(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "INSERT INTO user (id, name, email) VALUES ($1, $2, $3)"
	wantArgs := []any{1, "Alice", "alice@example.com"}

	s := structsql.New() // Default PostgreSQL
	var gotSQL string
	gotArgs := make([]any, 0, 10) // Pre-allocate with capacity

	err := s.Insert(u, &gotSQL, &gotArgs)
	if err != nil {
		t.Fatalf("Insert error: %v", err)
	}

	if gotSQL != wantSQL {
		t.Fatalf("Insert SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("Insert args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
	}
}

func TestInsertSQLite(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "INSERT INTO user (id, name, email) VALUES (?, ?, ?)"
	wantArgs := []any{1, "Alice", "alice@example.com"}

	s := structsql.New(structsql.SQLite)
	var gotSQL string
	gotArgs := make([]any, 0, 10) // Pre-allocate with capacity

	err := s.Insert(u, &gotSQL, &gotArgs)
	if err != nil {
		t.Fatalf("Insert error: %v", err)
	}

	if gotSQL != wantSQL {
		t.Fatalf("Insert SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("Insert args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
	}
}

func BenchmarkInsert(b *testing.B) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	s := structsql.New()
	var sql string
	args := make([]any, 0, 10) // Pre-allocate with capacity
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args = args[:0] // Clear for reuse
		_ = s.Insert(u, &sql, &args)
	}
}

func BenchmarkInsertWithArgs(b *testing.B) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	s := structsql.New()
	var sql string
	args := make([]any, 0, 10) // Pre-allocate with capacity
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		args = args[:0] // Clear for reuse (more efficient than nil assignment)
		_ = s.Insert(u, &sql, &args)
	}
}
