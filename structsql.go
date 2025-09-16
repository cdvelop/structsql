package structsql

import (
	. "github.com/cdvelop/tinystring"
)

type StructNamer interface {
	StructName() string
}

type FieldInfo struct {
	Name string
}

type TypeInfo struct {
	fields []FieldInfo
}

var typeCache = make(map[uintptr]*TypeInfo)

type Structsql struct{}

func New() *Structsql {

     s := &Structsql{}

     return s
}

func init() {
     // Pre-warm Conv pool to reduce allocations
     for i := 0; i < 10; i++ {
          c := GetConv()
          c.PutConv()
     }
}
