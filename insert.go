package structsql

import (
	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
	"unsafe"
)

func (s *Structsql) Insert(sql *string, values *[]any, structs ...any) error {
	if len(structs) == 0 {
		return Err("no structs provided")
	}

	// For now, handle only single struct (first one)
	v := structs[0]

	// Check if implements StructNamer
	namer, ok := v.(StructNamer)
	if !ok {
		return Err("struct does not implement StructNamer interface")
	}

	typ := tinyreflect.TypeOf(v)
	if typ.Kind() != 25 {
		return Err("input is not a struct")
	}

	// Get Conv for zero-allocation
	c := GetConv()
	defer c.PutConv()

	// Table name: StructName() lowercased + "s"
	tableName := namer.StructName()
	c.WrString(BuffOut, tableName)
	c.ToLower()
	c.WrString(BuffOut, "s")
	tableStr := c.GetString(BuffOut)
	c.ResetBuffer(BuffOut)

	// Reset for reuse
	c.ResetBuffer(BuffOut)

	// Get cached type info (slice-based lookup)
	typPtr := uintptr(unsafe.Pointer(typ))
	var typeInfo *TypeInfo

	// Find existing cache entry
	for _, entry := range s.typeCache {
		if entry.typePtr == typPtr {
			typeInfo = entry.info
			break
		}
	}

	if typeInfo == nil {
		// Build cache
		numFields, err := typ.NumField()
		if err != nil {
			return err
		}
		fields := make([]FieldInfo, numFields)
		for i := 0; i < numFields; i++ {
			field, err := typ.Field(i)
			if err != nil {
				return err
			}
			c := GetConv()
			c.WrString(BuffOut, field.Name.Name())
			c.ToLower()
			name := c.GetString(BuffOut)
			c.ResetBuffer(BuffOut)
			c.PutConv()
			fields[i] = FieldInfo{Name: name}
		}
		typeInfo = &TypeInfo{fields: fields}

		// Add to cache
		if len(s.typeCache) < cap(s.typeCache) {
			s.typeCache = append(s.typeCache, typeCacheEntry{typePtr: typPtr, info: typeInfo})
		}
		// If cache is full, don't cache (simple approach)
	}

	numFields := len(typeInfo.fields)
	if numFields == 0 {
		return Err("struct has no fields")
	}

	// Collect columns for SQL building
	var columns [32]string
	var colCount int

	for i := 0; i < numFields; i++ {
		fieldName := typeInfo.fields[i].Name
		columns[colCount] = fieldName
		colCount++
	}

	// Build SQL
	c.WrString(BuffOut, "INSERT INTO ")
	c.WrString(BuffOut, tableStr)
	c.WrString(BuffOut, " (")

	// Columns
	for i := 0; i < colCount; i++ {
		if i > 0 {
			c.WrString(BuffOut, ", ")
		}
		c.WrString(BuffOut, columns[i])
	}

	c.WrString(BuffOut, ") VALUES (")

	// Placeholders
	for i := 0; i < colCount; i++ {
		if i > 0 {
			c.WrString(BuffOut, ", ")
		}
		c.WrString(BuffOut, "?")
	}

	c.WrString(BuffOut, ")")

	*sql = c.GetStringZeroCopy(BuffOut)

	// Populate values slice (reuse caller's buffer)
	*values = (*values)[:0] // Clear existing values
	val := tinyreflect.ValueOf(v)
	for i := 0; i < numFields; i++ {
		fieldVal, err := val.Field(i)
		if err != nil {
			return err
		}

		iface, err := fieldVal.Interface()
		if err != nil {
			return err
		}

		*values = append(*values, iface) // Append to caller's buffer
	}

	return nil
}
