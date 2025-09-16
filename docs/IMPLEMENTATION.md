# Implementation Plan for Insert Function in StructSQL

## Overview
Implement the `Insert` function in `insert.go` to generate SQL INSERT statements from Go structs, adhering to the constraints of using only `tinystring` for string handling and `tinyreflect` for structure reflection.

## Constraints
- **Zero Memory Allocation**: Implementation must be optimized for zero memory allocation to support tinygo compilation and embedded environments.
- **No Standard Library**: Do not use any standard Go library functions; rely solely on `tinystring` for strings, errors, and numbers.
- **Interface Requirement**: Structs must implement `StructNamer` interface for table name derivation.
- **Error Handling**: All CRUD methods return errors using `tinystring` error functions.

## Requirements
- Input: A struct instance that implements `StructName() string` interface (e.g., `User{ID: 1, Name: "Alice", Email: "alice@example.com"}`)
- Output: SQL string, slice of interface{} for values, and error
- Example: `"INSERT INTO users (id, name, email) VALUES (?, ?, ?)"`, `[]interface{}{1, "Alice", "alice@example.com"}`, nil
- Use struct tags `db:"column_name"` for column names, fallback to field name
- Table name derived from `v.StructName()` (lowercased + 's' for pluralization)
- All CRUD methods return error: Insert/Update/Delete return (string, []interface{}, error), Select returns (string, error)
- Handle only struct inputs implementing the interface; return error for invalid inputs

## Dependencies
- `tinyreflect`: For inspecting struct types, fields, and values
- `tinystring`: For all string operations, error handling, and number conversions (no standard library usage allowed)

## Implementation Steps
1. **Update Function Signatures**: Change all CRUD functions to return error (Insert/Update/Delete: (string, []interface{}, error), Select: (string, error))
2. **Define Interface**: Add `type StructNamer interface { StructName() string }` in structsql.go
3. **Update Test Struct**: Make User implement StructName() string method
4. **Update Tests**: Modify test calls to handle the new return values
5. **Analyze Requirements**: Understand expected SQL format and args from test cases
6. **Struct Inspection**: Use `tinyreflect.TypeOf` to get type, check if struct and implements interface
7. **Table Name Derivation**: Convert `v.StructName()` to lowercase and pluralize (e.g., "User" -> "users")
8. **Field Extraction**: Iterate fields, extract column names from `db` tags or field names
9. **SQL Building**: Construct INSERT statement using `tinystring` functions
10. **Value Collection**: Gather field values into slice
11. **Edge Cases**: Handle non-struct inputs, structs not implementing interface, empty structs
12. **Testing**: Run `structsql_test.go` to verify correctness

## Detailed Design
- Define interface: `type StructNamer interface { StructName() string }`
- Check if `v` implements `StructNamer`, else return error
- Use `tinyreflect.TypeOf(v)` to get type
- If `typ.Kind() != K.Struct`, return error
- Use `tinystring.Conv` (pooled) for zero-allocation string operations:
  - Get Conv from pool with `GetConv()`, defer `putConv()`
  - Table name: Write `v.StructName()` to buffer, `ToLower()`, append "s"
  - For each field: Parse `db` tag using buffer operations, extract column name
  - Build column list and placeholders by writing to buffers
  - Construct final SQL string from buffers
- Collect field values into `[]interface{}` (may require allocation, but minimize)
- Return sql (from buffer), values, nil or error

## Todo List
- [x] Analyze the Insert function requirements from structsql_test.go
- [x] Understand tinyreflect API for struct inspection
- [x] Understand tinystring API for string handling
- [x] Define StructNamer interface in structsql.go
- [x] Update all CRUD function signatures to return error
- [x] Update User struct in test to implement StructName()
- [x] Update test calls to handle new return values
- [x] Implement table name derivation from StructName() (pluralization)
- [x] Implement field extraction with db tags
- [x] Build INSERT SQL statement using tinystring Conv buffers
- [x] Collect field values for args
- [x] Handle edge cases (non-struct input, missing interface, empty structs)
- [x] Run tests to verify implementation
- [x] Create benchmark for Insert to verify allocations

## Benchmark Results
BenchmarkInsert-16    	 2607606	       451.2 ns/op	     352 B/op	      11 allocs/op

The implementation currently has 11 allocations per operation, primarily from:
- []string slice for column names
- []interface{} slice for field values
- Conv buffer operations (though pooled)

Further optimization needed to achieve zero allocations for tinygo compatibility.

## Optimization Plan for Zero Allocations

### Current Allocation Sources Analysis
Benchmark results show 11 allocations per operation:
- 352 B/op total
- Primary sources: []string slice, []interface{} slice, reflection operations

### Alternative Optimization Strategies (API Preservation)

#### 1. tinyreflect Allocation Review
- **Investigate tinyreflect.Field() and NumField()**: Check if these methods allocate memory during struct field iteration
- **Potential Fix**: Modify tinyreflect to use pre-allocated field descriptors or cache field metadata
- **Impact**: Could reduce allocations from reflection operations

#### 2. Slice Pool Implementation
- **Introduce sync.Pool for slices**: Create pools for []string and []interface{} of common sizes
- **Size-based pooling**: Pools for small (4 fields), medium (16 fields), large (64 fields) structs
- **Return pooled slices**: Instead of new allocations, get from pool and return after use
- **Expected reduction**: 4-6 allocations eliminated

#### 3. Buffer Pre-allocation Strategy
- **Pre-warm Conv pool**: Ensure sufficient Conv objects are created at startup
- **Fixed buffer sizes**: Use larger initial buffer capacities to prevent expansion
- **Pool sizing**: Calculate based on expected concurrent operations

#### 4. Value Collection Optimization
- **Minimize interface{} boxing**: Use type assertions or unsafe operations for known types
- **Deferred boxing**: Collect values as concrete types, box only when returning
- **But API constraint**: Must return []interface{}, so boxing unavoidable

#### 5. Compile-time Optimization
- **Code generation approach**: Generate type-specific Insert functions at compile time
- **Eliminate runtime reflection**: Use generated code that directly accesses fields
- **Zero runtime allocation**: All operations resolved at compile time
- **Trade-off**: Requires code generation tooling, changes development workflow

### Proposed Implementation Changes

1. **Slice Pooling**:
   ```go
   var stringSlicePool = sync.Pool{New: func() any { return make([]string, 0, 16) }}
   var interfaceSlicePool = sync.Pool{New: func() any { return make([]interface{}, 0, 16) }}
   ```
   - Get slices from pool, use append, return to pool

2. **tinyreflect Optimization**:
   - Review and optimize tinyreflect.Field() to avoid allocations
   - Cache field metadata per type

3. **Buffer Optimization**:
   - Increase default Conv buffer sizes
   - Pre-populate pool

### Expected Outcome
- Reduce allocations from 11 to 2-4
- Maintain exact API compatibility
- Improve performance for repeated calls
- Better tinygo compatibility

### Implementation Priority
1. Implement slice pooling for []string and []interface{}
2. Optimize Conv buffer management
3. Review tinyreflect for allocation opportunities
4. Consider compile-time code generation as future enhancement

## Next Steps
Await user approval before proceeding to code implementation in Code mode.