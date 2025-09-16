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
  - Get Conv from pool with `getConv()`, defer `putConv()`
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
- [ ] Implement table name derivation from struct name (pluralization)
- [ ] Implement field extraction with db tags
- [ ] Build INSERT SQL statement using tinystring
- [ ] Collect field values for args
- [ ] Handle edge cases (non-struct input, empty structs)
- [ ] Run tests to verify implementation

## Next Steps
Await user approval before proceeding to code implementation in Code mode.