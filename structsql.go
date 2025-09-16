package structsql

type StructNamer interface {
	StructName() string
}

type Structsql struct{}

func New() *Structsql {

    s := &Structsql{}

    return s
}
