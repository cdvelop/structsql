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
	convPool  *Conv
}

type typeCacheEntry struct {
	typePtr uintptr
	info    *TypeInfo
}

func New() *Structsql {
	// Get a Conv from pool but don't return it - keep it for this instance
	conv := GetConv()

	s := &Structsql{
		typeCache: make([]typeCacheEntry, 0, 16), // Pre-allocate capacity
		convPool:  conv,                          // Single Conv instance per Structsql
	}

	return s
}
