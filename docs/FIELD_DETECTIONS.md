# Field Detection for Partial Updates

## Implementation Status: ✅ COMPLETED

This document describes the implemented field detection system for partial updates in StructSQL, which automatically excludes zero-valued fields from UPDATE statements.

## Current Implementation Analysis

### Update Operation
- Currently includes ALL non-ID fields in SET clause
- SQL: `UPDATE user SET name=$1, email=$2 WHERE id=$3`
- Values: ["Alice", "alice@example.com", 1]
- No filtering of zero/unset fields

### Delete Operation
- Only uses ID field (already correct)
- SQL: `DELETE FROM user WHERE id=$1`
- Values: [1]

### Insert Operation
- Includes ALL fields (appropriate for inserts)
- No change needed

## Problem Statement

When providing partial data like:
```go
u := User{ID: 1, Email: "alice@example.com"} // Name is ""
```

Current Update generates:
```sql
UPDATE user SET name=$1, email=$2 WHERE id=$3
```
With values: ["", "alice@example.com", 1]

**Issue**: This sets name to empty string even when only email should be updated.

## Desired Behavior

For partial updates, only include non-zero fields in SET clause:
```sql
UPDATE user SET email=$1 WHERE id=$2
```
Values: ["alice@example.com", 1]

## Standard Library Reflect Approach

### Zero Value Detection
Go's standard `reflect` package provides `Value.IsZero()` method (since Go 1.13):

```go
func (v Value) IsZero() bool
```

**Implementation details:**
- **Basic types**: Direct comparison with zero value (`""`, `0`, `false`, etc.)
- **Pointers**: `nil` check
- **Slices/Maps**: `nil` or empty check
- **Structs**: Recursively checks if all fields are zero
- **Interfaces**: `nil` check

### ORM Patterns
Popular Go ORMs handle partial updates differently:

- **GORM**: Uses field tracking with `Select()` method or struct tags
- **sqlx**: Manual field specification
- **Standard approach**: Zero value filtering (like proposed here)

### Key Insights for TinyReflect
- `IsZero()` is fundamental and widely used
- Implementation should mirror standard library behavior
- Performance-critical for reflection-heavy code
- Essential for data serialization and ORM operations

## TinyReflect Analysis

### Current Capabilities
- `ValueOf()` - Create reflected value
- `Field(i)` - Access struct fields
- `Interface()` - Get value as interface{}
- `IsZero()` - Zero value detection ✅ **IMPLEMENTED**
- Basic type accessors: `String()`, `Int()`, `Bool()`, etc.

### Philosophy Alignment
TinyReflect maintains minimal API focused on essential operations. The IsZero method fits this philosophy as it's fundamental for data handling and aligns with standard library patterns.

## Current Implementation

### 1. TinyReflect IsZero Method

The `IsZero() bool` method is implemented in `tinyreflect/ValueOf.go`, following standard library patterns:

```go
// IsZero reports whether v is the zero value for its type.
// It mirrors reflect.Value.IsZero() behavior for supported types.
func (v Value) IsZero() bool {
    // Handle nil Value (from ValueOf(nil))
    if v.typ_ == nil {
        return true
    }

    switch v.kind() {
    case K.String:
        return *(*string)(v.ptr) == ""
    case K.Bool:
        return !*(*bool)(v.ptr)
    case K.Int:
        return *(*int)(v.ptr) == 0
    // ... (additional integer, float, pointer, slice, map, struct cases)
    default:
        return false
    }
}
```

**Key Features:**
- Mirrors standard library `reflect.Value.IsZero()` behavior
- Supports all primitive types, pointers, slices, maps, and structs
- Recursively checks struct fields
- Handles nil values appropriately

### 2. Update Function Implementation

The `update.go` implementation automatically filters zero fields:

```go
// Collect SET fields (non-zero, non-id)
var setColumns [32]string
var setCount int
for i := 0; i < numFields; i++ {
    if i != idIndex {
        fieldVal, err := val.Field(i)
        if err != nil {
            return err
        }
        if !fieldVal.IsZero() {
            setColumns[setCount] = info.fields[i].Name
            setCount++
        }
    }
}
```

### 3. Test Coverage

Comprehensive tests are implemented in `tinyreflect/ValueOf.IsZero_test.go` and `update_test.go`:

```go
func TestIsZero(t *testing.T) {
    // Tests for all supported types including primitives, structs, slices, maps
}

func TestUpdatePartial(t *testing.T) {
    u := User{ID: 1, Email: "alice@example.com"} // Name is zero
    wantSQL := "UPDATE user SET email=$1 WHERE id=$2"
    wantArgs := []any{"alice@example.com", 1}
    // ... assertions
}
```

## Edge Cases & Considerations

### Intentional Zero Values
- Problem: Cannot distinguish between "not provided" vs "set to zero"
- Current approach: Treat zero as "not provided" (common in ORMs)
- Alternative: Could add field tags or separate "provided" tracking

### Supported Types
- All tinyreflect-supported types need IsZero implementation
- Basic types: straightforward ("" == zero, 0 == zero, false == zero)
- Complex types: slices/maps zero when nil or empty?

### Performance Impact
- Minimal: IsZero check is fast for basic types
- Avoids unnecessary SQL parameters

### Backwards Compatibility
- Existing full-struct updates continue working
- Partial updates become possible

## Implementation Summary

✅ **All implementation steps completed:**

1. **IsZero method added to tinyreflect/ValueOf.go** - Full implementation with support for all types
2. **Update function modified in structsql/update.go** - Automatic zero field filtering
3. **Comprehensive tests added** - Both unit tests and integration tests
4. **Cross-database compatibility** - Works with PostgreSQL and SQLite

## Current Behavior

- **Zero values are treated as "not provided"** - This is the standard ORM approach
- **All tinyreflect-supported types** have IsZero implementation
- **No field-level tags required** - Zero value detection is automatic
- **Backward compatible** - Existing code continues to work

## Usage Example

```go
s := structsql.New()

// Partial update - only non-zero fields are included
user := User{ID: 1, Email: "new@example.com"} // Name remains ""
err := s.Update(user, &sql, &values)
// Generated: UPDATE user SET email=$1 WHERE id=$2
// Values: ["new@example.com", 1]
```