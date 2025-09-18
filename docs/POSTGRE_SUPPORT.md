# PostgreSQL and SQLite Support Proposal for StructSQL

## Overview
This document proposes adding support for PostgreSQL and SQLite database dialects to the StructSQL library, following the existing Insert function architecture. PostgreSQL will be the default database type, with SQLite as an alternative option. The implementation maintains the library's constraints: zero memory allocations, no standard library usage, and compatibility with tinygo.

## Current Architecture Constraints
- **Zero Memory Allocation**: Optimized for minimal heap allocations
- **No Standard Library**: Relies on tinystring and tinyreflect only
- **Tinygo Compatibility**: No unsafe operations
- **API Cleanliness**: User-facing variables are accessible but types remain private

## Proposed Changes

### 1. Database Type Definition
Define `dbType` as a custom string type with methods for placeholder generation:

```go
// Custom type based on string for type safety
type dbType string

// Public global variables for user access
const (
    PostgreSQL dbType = "postgres"
    SQLite     dbType = "sqlite"
)

// Placeholder method for each database type
func (d dbType) placeholder(index int, conv *Conv) {
    switch d {
    case PostgreSQL:
        conv.WrString(BuffOut, "$")
        conv.WrInt(BuffOut, index)
    case SQLite:
        conv.WrString(BuffOut, "?")
    }
}
```

### 2. Separate Configuration Files
Create dedicated files for each database's placeholder logic:

**ph_postgre.go**:
```go
package structsql

import . "github.com/cdvelop/tinystring"

func (d dbType) placeholderPostgre(index int, conv *Conv) {
    conv.WrString(BuffOut, "$")
    conv.WrInt(BuffOut, index)
}
```

**ph_sqlite.go**:
```go
package structsql

import . "github.com/cdvelop/tinystring"

func (d dbType) placeholderSQLite(index int, conv *Conv) {
    conv.WrString(BuffOut, "?")
}
```

### 3. Structsql Struct Enhancement
Add a `dbType` field to the `Structsql` struct with PostgreSQL as default:

```go
type Structsql struct {
    typeCache []typeCacheEntry
    convPool  *Conv
    dbType    dbType  // New field for database type
}
```

### 4. New() Function Modification
Modify the `New()` function to accept variadic arguments (`...any`) for future extensibility:

```go
func New(configs ...any) *Structsql {
    db := PostgreSQL  // Default to PostgreSQL

    // Parse configurations
    if len(configs) > 0 {
        if dt, ok := configs[0].(dbType); ok {
            db = dt
        }
    }

    conv := GetConv()
    s := &Structsql{
        typeCache: make([]typeCacheEntry, 0, 16),
        convPool:  conv,
        dbType:    db,
    }
    return s
}
```

### 5. Insert() Method Updates
Use the dbType method for placeholder generation:

```go
// In the placeholders generation section (around line 114-119)
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

## Future Considerations
- **Select Operations**: Select methods can adopt the same dbType logic for WHERE clauses
- **Additional Databases**: MySQL, SQL Server support can be added similarly
- **Advanced Configurations**: Connection strings, schema prefixes, etc., via additional `...any` parameters
- **Batch Operations**: Support for multiple records in single operations

## Migration Path
No breaking changes for existing users. New users can specify database type explicitly, while defaults ensure PostgreSQL compatibility.

## Testing Strategy
- Unit tests for both PostgreSQL and SQLite placeholder generation
- Benchmark tests to ensure no performance regression
- Integration tests with actual database drivers (if feasible within constraints)

This proposal maintains the library's performance characteristics while adding essential database dialect support.