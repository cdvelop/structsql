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

func TestInsert(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "INSERT INTO users (id, name, email) VALUES (?, ?, ?)"
	wantArgs := []interface{}{1, "Alice", "alice@example.com"}

	gotSQL, gotArgs := structsql.Insert(u)

	if gotSQL != wantSQL {
		t.Fatalf("Insert SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("Insert args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
	}
}

func TestUpdate(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "UPDATE users SET name=?, email=? WHERE id=?"
	wantArgs := []interface{}{"Alice", "alice@example.com", 1}

	gotSQL, gotArgs := structsql.Update(u)

	if gotSQL != wantSQL {
		t.Fatalf("Update SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("Update args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
	}
}

func TestSelect(t *testing.T) {
	wantSQL := "SELECT id, name, email FROM users WHERE id = ?"
	gotSQL := structsql.Select(User{})
	if gotSQL != wantSQL {
		t.Fatalf("Select SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}
}

func TestDelete(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	wantSQL := "DELETE FROM users WHERE id=?"
	wantArgs := []interface{}{1}

	gotSQL, gotArgs := structsql.Delete(u)

	if gotSQL != wantSQL {
		t.Fatalf("Delete SQL mismatch:\n got: %s\nwant: %s", gotSQL, wantSQL)
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("Delete args mismatch:\n got: %v\nwant: %v", gotArgs, wantArgs)
	}
}
