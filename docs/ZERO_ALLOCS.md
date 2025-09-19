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
