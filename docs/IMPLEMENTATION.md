# StructSQL Insert Function - Implementation Report

## Current API
```go
func (s *Structsql) Insert(structTable any, sql *string, values *[]any) error
```

## Architecture Improvements

### ✅ Instance-Based Design
- **Moved `typeCache` from global to Structsql field**: Better encapsulation and testability
- **Changed from map to slice**: Eliminates concurrency issues, reduces code complexity
- **Pre-allocated cache capacity**: 16 entries to minimize slice growth

### ✅ Constructor-Based Initialization
- **Moved Conv pool pre-warming to `New()`**: Eliminates `init()` function for better testability
- **Instance-level resource management**: Each Structsql instance manages its own resources
- **Predictable initialization**: Resources allocated at construction time

### ✅ Simplified Caching Strategy
- **Slice-based lookup**: O(n) lookup instead of O(1) map, but no sync complexity
- **Fixed capacity**: 16 cache entries, simple overflow handling
- **Per-instance caching**: Each Structsql maintains separate cache

## Key Features
- **Output Parameters by Reference**: SQL and values passed as pointers for intuitive usage
- **Method of StructSql**: Enables caching and state management
- **Variadic Arguments**: Supports multiple structs for batch operations
- **Tinyreflect Compatible**: Full generic type support
- **Tinygo Ready**: No unsafe operations required

## Usage Example
```go
s := structsql.New() // Defaults to PostgreSQL
var sql string
var values []any

err := s.Insert(user, &sql, &values)
// sql: "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)"
// values: [1, "Alice", "alice@example.com"]
```

// For SQLite support
s := structsql.New(structsql.SQLite)
err := s.Insert(user, &sql, &values)
// sql: "INSERT INTO users (id, name, email) VALUES (?, ?, ?)"

## Implementation Details

### Core Algorithm
1. **Type Validation**: Check StructNamer interface implementation
2. **SQL Generation**: Build INSERT statement using tinystring buffers
3. **Field Extraction**: Use tinyreflect to extract struct field values
4. **Value Population**: Populate caller's slice by reference

## Test Coverage
- ✅ Unit tests for SQL generation and value extraction
- ✅ Benchmark tests for performance validation
- ✅ Memory profiling for allocation analysis
- ✅ Edge case handling (empty structs, invalid types)