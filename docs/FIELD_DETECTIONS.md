# Field Detection for Partial Updates

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
- Basic type accessors: `String()`, `Int()`, `Bool()`, etc.
- **Missing**: Zero value detection

### Philosophy Alignment
TinyReflect maintains minimal API focused on essential operations. Adding zero detection fits this philosophy as it's fundamental for data handling and aligns with standard library patterns.

## Proposed Solution

### 1. Extend TinyReflect with IsZero Method

Add `IsZero() bool` method to `Value` struct, following standard library patterns:

```go
// IsZero reports whether v is the zero value for its type
// Mirrors reflect.Value.IsZero() behavior for supported types
func (v Value) IsZero() bool {
    switch v.kind() {
    case K.String:
        return *(*string)(v.ptr) == ""
    case K.Int:
        return *(*int)(v.ptr) == 0
    case K.Int8:
        return *(*int8)(v.ptr) == 0
    case K.Int16:
        return *(*int16)(v.ptr) == 0
    case K.Int32:
        return *(*int32)(v.ptr) == 0
    case K.Int64:
        return *(*int64)(v.ptr) == 0
    case K.Uint, K.Uint8, K.Uint16, K.Uint32, K.Uint64, K.Uintptr:
        return *(*uint64)(v.ptr) == 0 // Safe for all uint sizes
    case K.Float32:
        return *(*float32)(v.ptr) == 0
    case K.Float64:
        return *(*float64)(v.ptr) == 0
    case K.Bool:
        return !*(*bool)(v.ptr)
    case K.Pointer, K.Interface:
        return v.ptr == nil
    case K.Slice, K.Map:
        return v.ptr == nil
    case K.Struct:
        // Recursively check all fields
        num, _ := v.NumField()
        for i := 0; i < num; i++ {
            field, _ := v.Field(i)
            if !field.IsZero() {
                return false
            }
        }
        return true
    default:
        return false // Unknown types considered non-zero
    }
}
```

**Benefits:**
- Mirrors standard library `reflect.Value.IsZero()` behavior
- Enables automatic field detection for partial updates
- Reusable across applications maintaining tinyreflect's philosophy

### 2. Modify Update Function

Update `update.go` to filter zero fields:

```go
// Collect SET fields (non-zero, non-ID)
var setColumns [32]string
var setCount int
for i := 0; i < numFields; i++ {
    if i != idIndex {
        fieldVal, _ := val.Field(i)
        if !fieldVal.IsZero() {
            setColumns[setCount] = info.fields[i].Name
            setCount++
        }
    }
}
```

### 3. Test Updates

Add test cases for partial updates:
```go
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

## Implementation Steps

1. **Add IsZero to tinyreflect/ValueOf.go**
2. **Modify structsql/update.go** to use IsZero filtering
3. **Update tests** in update_test.go
4. **Test with both databases** (PostgreSQL/SQLite)

## Questions for Review

- Should zero values be treated as "not provided" or require explicit handling?
- Any special cases for specific field types?
- Need for field-level tags to control update behavior?