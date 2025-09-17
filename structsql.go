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

type Structsql struct {
	typeCache []typeCacheEntry
	convPool  []*Conv
}

type typeCacheEntry struct {
	typePtr uintptr
	info    *TypeInfo
}

func New() *Structsql {
	s := &Structsql{
		typeCache: make([]typeCacheEntry, 0, 16), // Pre-allocate capacity
		convPool:  make([]*Conv, 0, 10),          // Pre-allocate capacity
	}

	// Pre-warm Conv pool
	for i := 0; i < 10; i++ {
		c := GetConv()
		s.convPool = append(s.convPool, c)
	}

	return s
}
