package structsql_test

type User struct {
	ID    int    `db:"id,pk"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func (u User) StructName() string {
	return "User"
}
