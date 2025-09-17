package structsql_test

import (
	"reflect"
	"testing"

	"github.com/cdvelop/structsql"
)

type User struct {
	ID    int    `db:"id,pk"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func (u User) StructName() string {
	return "User"
}

func TestInsert(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "INSERT INTO users (id, name, email) VALUES (?, ?, ?)"
	wantArgs := []any{1, "Alice", "alice@example.com"}

	s := structsql.New()
	var gotSQL string
	var gotArgs []any

	err := s.Insert(&gotSQL, &gotArgs, u)
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
	var args []any
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Insert(&sql, &args, u)
	}
}

func BenchmarkInsertWithArgs(b *testing.B) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	s := structsql.New()
	var sql string
	var args []any
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Insert(&sql, &args, u)
		// Clear args for next iteration
		for j := range args {
			args[j] = nil
		}
	}
}

func TestUpdate(t *testing.T) {
	/*
		 	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
			wantSQL := "UPDATE users SET name=?, email=? WHERE id=?"
			wantArgs := []interface{}{"Alice", "alice@example.com", 1}

			gotSQL, gotArgs, err := structsql.Update(u)
			if err != nil {
				t.Fatalf("Update error: %v", err)
			}

			if gotSQL != wantSQL {
				t.Fatalf("Update SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
			}
			if !reflect.DeepEqual(gotArgs, wantArgs) {
				t.Fatalf("Update args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
			}
	*/
}

func TestSelect(t *testing.T) {
	/* wantSQL := "SELECT id, name, email FROM users WHERE id = ?"
	gotSQL, err := structsql.Select(User{})
	if err != nil {
		t.Fatalf("Select error: %v", err)
	}
	if gotSQL != wantSQL {
		t.Fatalf("Select SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	} */
}

func TestDelete(t *testing.T) {
	/* u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "DELETE FROM users WHERE id=?"
	wantArgs := []interface{}{1}

	gotSQL, gotArgs, err := structsql.Delete(u)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	if gotSQL != wantSQL {
		t.Fatalf("Delete SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("Delete args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
	} */
}
