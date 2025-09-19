# Zero Allocations in Go with structsql

## Analysis Tools

```bash
# Identify allocations at compile time
go test -print-allocs=.

# Memory profiling
go test -benchmem -bench=.
```

## Overview
High-performance SQL INSERT generation for Go structs using tinyreflect and tinystring, optimized for zero memory allocations and tinygo compatibility.

## Architecture Constraints
- **Zero Memory Allocation**: Implementation optimized for minimal heap allocations to support tinygo compilation and embedded environments
- **No Standard Library**: Cannot use any standard Go library functions
- **Allowed Libraries**: Only `tinystring` for string/errors/numbers operations and `tinyreflect` for type reflection
- **Interface Requirement**: Structs must implement `StructNamer` interface for table name derivation
- **Error Handling**: All methods return errors using `tinystring` error functions

## Current Library Status

### Insert Operation
- **Benchmark Results**: 52 B/op, 2 allocs/op, ~181.9 ns/op
- **Allocation Sources**:
  - 1 alloc: `fieldVal.Interface()` boxing for struct field values (3 fields: ID, Name, Email)
  - 1 alloc: Internal slice growth or reflection overhead

### Update Operation
- **Benchmark Results**: 52 B/op, 2 allocs/op, ~252.6 ns/op
- **Allocation Sources**:
  - 1 alloc: `fieldVal.Interface()` boxing for non-zero SET fields (Name, Email) and ID field
  - 1 alloc: Internal slice growth or reflection overhead

### Delete Operation
- **Benchmark Results**: 52 B/op, 2 allocs/op, ~177.2 ns/op
- **Allocation Sources**:
  - 1 alloc: `fieldVal.Interface()` boxing for ID field
  - 1 alloc: Internal slice growth or reflection overhead

## Performance Results (Historical - Insert Only)

### Benchmark Results (Latest - Profiling Validated)
- **Memory Usage**: 48 B/op (**92% reduction** from 624 B/op)
- **Performance**: ~138.9 ns/op (**69% improvement** from ~450 ns/op)
- **Allocations**: 1 allocs/op (**67% reduction** from 3 allocs/op)
- **Cache Strategy**: Slice-based (16 entries capacity)
- **Initialization**: Constructor-based (instance-level Conv)
- **GetConv() Calls**: **Eliminated** (0 calls)

### Current Status
- **✅ GetConv() Eliminated**: Single Conv instance per Structsql (0 pool calls)
- **✅ Performance Improved**: 5% boost from instance-based Conv
- **✅ Memory Optimized**: 52 B/op stable
- **⚠️ Remaining**: 2 allocs from interface{} boxing and internal overhead (52 B/op)

## Memory Optimizations
- **Type Caching**: Cache struct metadata per type
- **Buffer Pooling**: Reuse Conv buffers for string operations
- **Reference Parameters**: Avoid return value allocations
- **Fixed Arrays**: Pre-allocated arrays for intermediate storage

## Key Findings from Memory Profiling
- **91% Memory Reduction**: From 624 B/op to 52 B/op achieved
- **33% Allocation Reduction**: From 3 allocs/op to 2 allocs/op
- **60% Performance Improvement**: From ~450 ns/op to ~181 ns/op
- **Primary Allocation Eliminated**: GetConv() pool calls (0 calls remaining)
- **Remaining Allocations**: `fieldVal.Interface()` boxing and internal overhead (52 B/op)

## Final Implementation Status

The StructSQL library has been optimized to achieve the best possible performance within the constraints of the current API design. All operations (Insert, Update, Delete) maintain 2 allocations per operation due to the required `[]interface{}` output format and internal reflection overhead.

### Key Achievements
- **91% Memory Reduction**: From 624 B/op to 52 B/op
- **33% Allocation Reduction**: From 3 allocs/op to 2 allocs/op
- **60% Performance Improvement**: From ~450 ns/op to ~181-253 ns/op
- **Zero GetConv() Calls**: Eliminated pool overhead through instance-based Conv management
- **Full TinyGo Compatibility**: No unsafe operations or standard library dependencies

### Technical Constraints
The remaining 2 allocations per operation are caused by:
1. `fieldVal.Interface()` boxing when extracting struct field values for dynamic storage in `[]any` slices
2. Internal reflection overhead from tinyreflect operations

These allocations are unavoidable with the current API design, as Go's interface{} boxing is required for dynamic type storage in slices, and reflection inherently involves some overhead.

### Future Considerations
Any further optimizations would require API changes (e.g., generics or callback-based approaches) that break backward compatibility. The current implementation represents the optimal balance of performance, compatibility, and maintainability within the constraints of Go's reflection system and the required `[]any` output format.

## PLAN: Zero Allocations Strategy

Based on comprehensive analysis and memory profiling, I have identified **specific sources** of the remaining 2 allocations and designed **concrete strategies** to eliminate them completely.

### Current Allocation Analysis (September 2025)

**Memory Profile Results:**
```
Type: alloc_objects
Showing nodes accounting for 8367102, 99.95% of 8371407 total
      flat  flat%   sum%        cum   cum%
   6958057 83.12% 83.12%    8367102 99.95%  BenchmarkInsert
   1409045 16.83% 99.95%    1409045 16.83%  (*Conv).GetString (inline)
```

**Confirmed Allocation Sources:**
1. **String Allocation in `getTableName()`**: `GetString()` call in `shared.go:42` (16.83% of objects)
2. **Interface Boxing in Value Extraction**: `fieldVal.Interface()` calls for struct field values

### PHASE 1: Immediate Optimizations (0 Breaking Changes)

#### Strategy 1.1: Eliminate String Allocation in Table Name Generation
**Problem**: `shared.go:42` uses `GetString()` instead of `GetStringZeroCopy()`
```go
// Current (1 allocation):
*tableStr = c.GetString(BuffOut)

// Solution (0 allocations):
*tableStr = c.GetStringZeroCopy(BuffOut)
```

**Implementation:**
- Replace `GetString()` with `GetStringZeroCopy()` in `getTableName()`
- Replace `GetString()` with `GetStringZeroCopy()` in field name processing (`shared.go:70`)
- **Expected Reduction**: -1 allocation per operation

#### Strategy 1.2: Implement Zero-Alloc Interface Extraction
**Problem**: `fieldVal.Interface()` creates interface{} boxing
```go
// Current (1 allocation per field):
iface, err := fieldVal.Interface()
*values = append(*values, iface)

// Solution (0 allocations for primitives):
var iface any
fieldVal.InterfaceZeroAlloc(&iface)
*values = append(*values, iface)
```

**Implementation:**
- Use `InterfaceZeroAlloc()` method from tinyreflect v0.8.1+
- This eliminates boxing for primitive types (int, string, bool, float64, etc.)
- **Expected Reduction**: -1 allocation per operation

### PHASE 2: Advanced Optimizations (Potential Breaking Changes)

#### Strategy 2.1: Pre-allocated Value Containers
**Concept**: Use fixed-size arrays for values instead of dynamic slices
```go
// Current API:
func (s *Structsql) Insert(structTable any, sql *string, values *[]any) error

// Alternative API (breaking):
func (s *Structsql) InsertFixed(structTable any, sql *string, values *[8]any, count *int) error
```

#### Strategy 2.2: Callback-Based Value Extraction
**Concept**: Avoid []any slice creation entirely
```go
// Alternative API (breaking):
type ValueCallback func(index int, value any)
func (s *Structsql) InsertCallback(structTable any, sql *string, callback ValueCallback) error
```

#### Strategy 2.3: Generic Type-Safe API
**Concept**: Use Go generics to eliminate interface{} entirely
```go
// Alternative API (breaking):
func Insert[T StructNamer](s *Structsql, data T) (string, []any, error)
```

### PHASE 3: Implementation Roadmap

#### Step 3.1: Phase 1 Implementation (Immediate - 0 allocations target)
1. **Fix String Allocations (Week 1)**
   - Update `shared.go:42`: `GetString()` → `GetStringZeroCopy()`
   - Update `shared.go:70`: `GetString()` → `GetStringZeroCopy()` 
   - Test: Expect ~1 allocation reduction

2. **Implement Zero-Alloc Interface Extraction (Week 1)**
   - Replace all `fieldVal.Interface()` calls with `fieldVal.InterfaceZeroAlloc(&target)`
   - Update insert.go, update.go, delete.go
   - Test: Expect final allocation elimination

3. **Validation (Week 1)**
   - Run full benchmark suite
   - Confirm 0 B/op, 0 allocs/op across all operations
   - Validate TinyGo compatibility

#### Step 3.2: Phase 2 Research (Future)
- Design breaking change APIs for maximum performance
- Community feedback on API design preferences
- Compatibility layer considerations

### Technical Implementation Details

#### String Zero-Copy Fix
```go
// In getTableName() - shared.go:42
func (s *Structsql) getTableName(typ *tinyreflect.Type, tableStr *string) {
    c := s.convPool
    tableName := typ.Name()
    c.WrString(BuffOut, tableName)
    c.ToLower()
    *tableStr = c.GetStringZeroCopy(BuffOut)  // ← FIXED: was GetString()
    c.ResetBuffer(BuffOut)
}
```

#### Interface Zero-Alloc Fix
```go
// In insert.go, update.go, delete.go
for i := 0; i < numFields; i++ {
    fieldVal, err := val.Field(i)
    if err != nil {
        return err
    }
    
    var iface any
    fieldVal.InterfaceZeroAlloc(&iface)  // ← FIXED: was fieldVal.Interface()
    *values = append(*values, iface)
}
```

### Success Metrics

**Target Performance (Phase 1 Completion):**
- **Memory Usage**: 0 B/op (**100% reduction** from 52 B/op)
- **Allocations**: 0 allocs/op (**100% reduction** from 2 allocs/op)
- **Performance**: ~150-200 ns/op (maintained or improved)
- **Compatibility**: 100% backward compatible API

**Validation Criteria:**
```bash
# Expected benchmark results:
BenchmarkInsert-16        10000000    150.0 ns/op    0 B/op    0 allocs/op
BenchmarkUpdate-16         8000000    200.0 ns/op    0 B/op    0 allocs/op  
BenchmarkDelete-16        12000000    120.0 ns/op    0 B/op    0 allocs/op
```

### Risk Assessment

**Phase 1 Risks: MINIMAL**
- String zero-copy: Same API, different internal implementation
- InterfaceZeroAlloc: Proven method in tinyreflect, maintains same behavior
- No breaking changes to public API

**Phase 2 Risks: HIGH**
- Breaking API changes require major version bump
- Community adoption considerations
- Migration complexity for existing users

### Conclusion

**Phase 1 provides a clear path to achieving true zero allocations** without breaking changes by:
1. Fixing identified string allocation leak
2. Leveraging existing InterfaceZeroAlloc capability

This plan transforms StructSQL from **2 allocs/op to 0 allocs/op** while maintaining 100% API compatibility and TinyGo support.
