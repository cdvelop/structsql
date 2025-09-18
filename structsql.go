package structsql

import (
	. "github.com/cdvelop/tinystring"
)

// dbType represents database types for SQL generation
type dbType string

// Database type constants
const (
	PostgreSQL dbType = "postgres"
	SQLite     dbType = "sqlite"
)

// placeholder generates the appropriate placeholder for the database type
func (d dbType) placeholder(index int, conv *Conv) {
	switch d {
	case PostgreSQL:
		placeholderPostgre(index, conv)
	case SQLite:
		placeholderSQLite(index, conv)
	}
}

type fieldInfo struct {
	Name string
}

type typeInfo struct {
	fields []fieldInfo
}

type Structsql struct {
	typeCache []typeCacheEntry
	convPool  *Conv
	dbType    dbType
}

type typeCacheEntry struct {
	typePtr uintptr
	info    *typeInfo
}

func New(configs ...any) *Structsql {
	db := PostgreSQL // Default to PostgreSQL

	// Parse configurations
	if len(configs) > 0 {
		if dt, ok := configs[0].(dbType); ok {
			db = dt
		}
	}

	// Get a Conv from pool but don't return it - keep it for this instance
	conv := GetConv()

	s := &Structsql{
		typeCache: make([]typeCacheEntry, 0, 16), // Pre-allocate capacity
		convPool:  conv,                          // Single Conv instance per Structsql
		dbType:    db,
	}

	return s
}
