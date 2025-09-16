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

## Implementation Status
- [x] Analyze requirements and design API
- [x] Implement core Insert functionality with tinyreflect/tinystring
- [x] Add StructNamer interface requirement
- [x] Update all CRUD function signatures for error handling
- [x] Implement zero-allocation optimizations (Conv buffers, fixed arrays)
- [x] Create benchmarks and verify performance
- [x] Achieve 45% allocation reduction (11 → 6 allocs/op)
- [ ] Implement advanced optimizations for zero allocations

## Benchmark Results

### Before Optimization
BenchmarkInsert-16    	 2607606	       451.2 ns/op	     352 B/op	      11 allocs/op

### After Fixed-Array Optimization
BenchmarkInsert-16    	 3588081	       329.6 ns/op	     640 B/op	       6 allocs/op

**Improvement**: Allocations reduced from 11 to 6 (45% reduction), performance improved from 451.2 ns/op to 329.6 ns/op.

### After Type Caching Optimization
BenchmarkInsert-16    	 4659055	       261.3 ns/op	     624 B/op	       3 allocs/op

**Improvement**: Allocations reduced from 6 to 3 (50% reduction), performance improved from 329.6 ns/op to 261.3 ns/op.

### Analysis of Remaining 6 Allocations

Detailed breakdown of remaining allocations based on code analysis:

1. **Conv Pool Allocation (1-2 allocs)**: `GetConv()` may allocate if pool is exhausted
2. **Slice Header Creation (1 alloc)**: `values[:valCount]` creates slice header for return
3. **Reflection Operations (2-3 allocs)**:
   - `tinyreflect.TypeOf(v)` - type resolution
   - `typ.Field(i)` - field descriptor creation
   - `tinyreflect.ValueOf(v)` - value wrapper
   - `fieldVal.Interface()` - interface boxing
4. **Buffer Operations (0-1 alloc)**: Potential buffer expansion in Conv

### New Optimization Plan (API Preservation)

#### Phase 1: Type Information Caching
- Implement global cache for struct type metadata
- Cache field names, types, and tag information per struct type
- Avoid repeated reflection.Field() calls
- **Expected reduction**: 1-2 allocs from reflection

#### Phase 2: Unsafe Value Extraction
- Use unsafe.Pointer operations to access struct fields directly
- Bypass interface{} boxing for primitive types
- Implement type-specific value extraction functions
- **Expected reduction**: 2 allocs from boxing

#### Phase 3: Buffer Pre-allocation and Reuse
- Pre-allocate Conv buffers at package init
- Ensure pool never exhausts during normal operation
- Optimize buffer size calculations
- **Expected reduction**: 1 alloc from pool exhaustion

#### Phase 4: Stack-Based Operations
- Use stack-allocated arrays for intermediate operations
- Minimize heap allocations by using local variables
- Optimize string building to use single buffer pass
- **Expected reduction**: 1 alloc from temporary objects

#### Phase 5: Inline Optimizations
- Inline critical path operations
- Eliminate function call overhead in hot paths
- Use compile-time optimizations where possible
- **Expected reduction**: 0-1 alloc

### Implementation Strategy
1. **Type Cache Implementation**:
   ```go
   var typeCache = make(map[uintptr]*TypeInfo)
   type TypeInfo struct {
       fields []FieldInfo
   }
   ```

2. **Unsafe Field Access**:
   - Calculate field offsets at runtime
   - Use unsafe.Pointer arithmetic for value extraction
   - Maintain type safety through careful offset calculations

3. **Buffer Pool Optimization**:
   - Increase pool size
   - Pre-warm with multiple Conv objects
   - Monitor pool usage in benchmarks

4. **Stack Allocation**:
   - Use [32]string and [32]interface{} on stack
   - Avoid heap allocation for small structs

### New Proposal: Unsafe Offset Calculation with Tinyreflect

#### Overview
Calculate field offsets using tinyreflect (once per type) by taking addresses of field values relative to struct pointer, cache the offsets, then use unsafe.Pointer arithmetic for direct field access at runtime. This eliminates reflection boxing while maintaining the generic API.

#### Implementation Strategy

1. **Offset Calculation**:
   ```go
   // In cache building
   val := tinyreflect.ValueOf(v)
   fieldVal, _ := val.Field(i)
   // Assume tinyreflect.Value has Addr() method
   fieldAddr := fieldVal.Addr().Pointer()
   structAddr := uintptr(unsafe.Pointer(&v))
   offset := fieldAddr - structAddr
   ```

2. **Runtime Access**:
   ```go
   // In value extraction
   ptr := unsafe.Pointer(&v)
   switch field.Kind {
   case 2: // int
       val := *(*int)(unsafe.Pointer(uintptr(ptr) + offset))
       values[valCount] = val
   }
   ```

3. **Cache Storage**:
   Extend FieldInfo to include offset and kind from tinyreflect.

#### Benefits
- **Zero Allocations**: Direct unsafe access, no reflection boxing
- **Generic API**: Maintains tinyreflect for type inspection
- **Performance**: Fast unsafe access after initial calculation
- **Tinygo Compatible**: Unsafe operations work in constrained environments

#### Assumptions
- tinyreflect.Value has Addr() and Pointer() methods (similar to Go reflect)
- Field layout is consistent across instances
- Memory alignment is handled properly

#### Expected Results
- **Target**: 0 allocs/op (slice header may remain)
- **Performance**: <250 ns/op
- **Compatibility**: Full tinygo support

### Safety Considerations
- Unsafe operations require careful validation
- Type safety must be maintained
- Bounds checking for array access
- Memory alignment considerations


### Implementation Roadmap
1. Implement type information caching
2. Add unsafe offset calculation with tinyreflect (new proposal)
3. Implement unsafe value extraction at runtime
4. Optimize buffer pool management
5. Apply stack-based optimizations
6. Target: Achieve 0 allocs for tinygo compatibility

### Expected Final Results
- **Target**: 0 allocs/op
- **Performance**: <250 ns/op
- **Compatibility**: Full tinygo support

### Important Notes
- **ToLower Usage**: Always use `Conv.ToLower()` method from `tinystring/capitalize.go`, not standalone functions
- **Buffer Operations**: All string processing must go through `Conv` pooled buffers for zero-allocation
- **API Preservation**: All optimizations maintain the existing function signatures

## Optimization Plan for Zero Allocations

### Current Allocation Sources Analysis
Benchmark results show 11 allocations per operation:
- 352 B/op total
- Primary sources: []string slice header, []interface{} slice header, dynamic slice growth, reflection operations

### Fixed-Size Array Strategy (API Preservation)

#### Core Concept
- Use fixed-size arrays instead of dynamic slices for column names and values
- Assume maximum struct fields (e.g., 32) and pre-allocate arrays
- Return slices of used portion: array[:count]
- Eliminates dynamic slice allocation and growth

#### Implementation Details

1. **Fixed Arrays**:
   ```go
   var columns [32]string
   var values [32]interface{}
   var colCount, valCount int
   ```

2. **Append Operation**:
   ```go
   columns[colCount] = fieldName
   colCount++
   values[valCount] = iface
   valCount++
   ```

3. **Return Slices**:
   ```go
   return sql, values[:valCount], nil
   ```

#### Benefits
- **No dynamic allocation**: Arrays are fixed size, allocated at function start
- **Minimal slice headers**: Only small slice headers for return values
- **Predictable memory usage**: No slice growth reallocations
- **Tinygo compatible**: Fixed sizes work better with constrained environments

#### Trade-offs
- **Size limit**: Maximum 32 fields per struct
- **Memory waste**: Unused array elements
- **Stack allocation**: Arrays on stack may increase stack usage

#### Expected Outcome
- Reduce allocations from 11 to 3-5 (slice headers + any buffer operations)
- Maintain exact API compatibility
- Enable tinygo compilation
- Predictable performance

### Alternative: Dynamic Pooling with Caller Responsibility
If fixed arrays are insufficient, implement pooling where caller manages slice lifecycle:

```go
func Insert(v any, columns *[]string, values *[]interface{}) (string, error) {
    // Append to provided slices
    *columns = append(*columns, fieldName)
    *values = append(*values, iface)
    return sql, nil
}
```

- Caller provides pre-allocated or pooled slices
- Zero allocations in Insert function
- Requires API change (not preferred)

### Implementation Priority
1. Implement fixed-size array approach
2. Test with benchmark
3. If insufficient, consider API change or other strategies

## Next Steps
New proposal: Unsafe offset calculation with tinyreflect for zero allocations.
Await user approval before proceeding to implementation.