# PostgreSQL and SQLite Support for StructSQL

## Implementation Status: ✅ COMPLETED

## Overview
StructSQL supports both PostgreSQL and SQLite database dialects, following the existing Insert function architecture. PostgreSQL is the default database type, with SQLite as an alternative option. The implementation maintains the library's constraints: zero memory allocations, no standard library usage, and compatibility with tinygo.

## Current Architecture Constraints
- **Zero Memory Allocation**: Optimized for minimal heap allocations
- **No Standard Library**: Relies on tinystring and tinyreflect only
- **Tinygo Compatibility**: No unsafe operations
- **API Cleanliness**: User-facing variables are accessible but types remain private

## Current Implementation

### 1. Database Type Definition
The `dbType` is defined as a custom string type with methods for placeholder generation:

```go
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
```

### 2. Separate Configuration Files
Dedicated files implement each database's placeholder logic:

**ph_postgre.go**:
```go
package structsql

import . "github.com/cdvelop/tinystring"

// placeholderPostgre generates PostgreSQL-style placeholders ($1, $2, ...)
func placeholderPostgre(index int, conv *Conv) {
    conv.WrString(BuffOut, "$")
    // Use AnyToBuff for tested integer-to-string conversion
    conv.AnyToBuff(BuffOut, index)
}
```

**ph_sqlite.go**:
```go
package structsql

import . "github.com/cdvelop/tinystring"

// placeholderSQLite generates SQLite-style placeholders (?, ?, ...)
func placeholderSQLite(index int, conv *Conv) {
    conv.WrString(BuffOut, "?")
}
```

### 3. Structsql Struct Enhancement
The `Structsql` struct includes a `dbType` field with PostgreSQL as default:

```go
type Structsql struct {
    typeCache []typeCacheEntry
    convPool  *Conv
    dbType    dbType
}
```

### 4. New() Function Implementation
The `New()` function accepts variadic arguments for configuration:

```go
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
```

### 5. SQL Generation Integration
All SQL generation methods use the dbType for placeholder generation:

```go
// In insert.go, update.go, delete.go
for i := 0; i < colCount; i++ {
    if i > 0 {
        c.WrString(BuffOut, ", ")
    }
    s.dbType.placeholder(i+1, c)  // Type-safe method call
}
```

## Usage Examples

### Default PostgreSQL Usage
```go
s := structsql.New()
// INSERT: INSERT INTO users (id, name, email) VALUES ($1, $2, $3)
// UPDATE: UPDATE users SET name=$1, email=$2 WHERE id=$3
// DELETE: DELETE FROM users WHERE id=$1
```

### Explicit PostgreSQL Configuration
```go
s := structsql.New(structsql.PostgreSQL)
// Same as default
```

### SQLite Configuration
```go
s := structsql.New(structsql.SQLite)
// INSERT: INSERT INTO users (id, name, email) VALUES (?, ?, ?)
// UPDATE: UPDATE users SET name=?, email=? WHERE id=?
// DELETE: DELETE FROM users WHERE id=?
```

### Complete API Usage
```go
s := structsql.New()

// Insert
var sql string
var values []any
err := s.Insert(user, &sql, &values)

// Update
err = s.Update(user, &sql, &values)

// Delete
err = s.Delete(user, &sql, &values)
```

## Implementation Benefits
- **Backward Compatibility**: Existing code continues to work with PostgreSQL defaults
- **Extensible API**: `...any` allows future configuration options
- **Type Safety**: Private types prevent API pollution while allowing user access
- **Performance**: Minimal overhead from database type checking
- **Consistency**: Follows existing architecture patterns from IMPLEMENTATION.md

## Current Status
- **All Operations**: Insert, Update, and Delete methods support both PostgreSQL and SQLite
- **Consistent API**: Same interface for both database types
- **Performance**: Minimal overhead from database type checking
- **Extensible**: `...any` parameters allow future configuration options

## Future Considerations
- **Additional Databases**: MySQL, SQL Server support can be added similarly
- **Advanced Configurations**: Connection strings, schema prefixes, etc., via additional `...any` parameters
- **Batch Operations**: Support for multiple records in single operations

## Migration Path
✅ **No breaking changes for existing users.** Existing code continues to work with PostgreSQL defaults. New users can specify database type explicitly.

## Testing Implementation
- ✅ Unit tests for both PostgreSQL and SQLite placeholder generation
- ✅ Benchmark tests to ensure no performance regression
- ✅ Integration tests with actual database drivers (within constraints)

The implementation maintains the library's performance characteristics while providing essential database dialect support.