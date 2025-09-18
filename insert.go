package structsql

import (
	"unsafe"

	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

func (s *Structsql) Insert(structTable any, sql *string, values *[]any) error {
	if structTable == nil {
		return Err("no struct table provided")
	}

	// For now, handle only single struct (first one)
	v := structTable

	typ := tinyreflect.TypeOf(v)
	if typ.Kind() != K.Struct {
		return Err("input is not a struct")
	}

	if typ.Name() == "struct" {
		return Err("struct does not implement StructNamer interface")
	}

	// Use instance Conv (no allocation)
	c := s.convPool
	c.ResetBuffer(BuffOut)
	c.ResetBuffer(BuffWork)
	c.ResetBuffer(BuffErr)

	// Table name: StructName() + "s" lowercased
	tableName := typ.Name() + "s"
	c.WrString(BuffOut, tableName)
	c.ToLower()
	tableStr := c.GetString(BuffOut)
	c.ResetBuffer(BuffOut)

	// Reset for reuse
	c.ResetBuffer(BuffOut)

	// Get cached type info (slice-based lookup)
	typPtr := uintptr(unsafe.Pointer(typ))
	var info *typeInfo

	// Find existing cache entry
	for _, entry := range s.typeCache {
		if entry.typePtr == typPtr {
			info = entry.info
			break
		}
	}

	if info == nil {
		// Build cache
		numFields, err := typ.NumField()
		if err != nil {
			return err
		}
		fields := make([]fieldInfo, numFields)
		for i := 0; i < numFields; i++ {
			field, err := typ.Field(i)
			if err != nil {
				return err
			}
			// Use instance Conv for field name processing
			s.convPool.WrString(BuffOut, field.Name.Name())
			s.convPool.ToLower()
			name := s.convPool.GetString(BuffOut)
			s.convPool.ResetBuffer(BuffOut)
			fields[i] = fieldInfo{Name: name}
		}
		info = &typeInfo{fields: fields}

		// Add to cache
		if len(s.typeCache) < cap(s.typeCache) {
			s.typeCache = append(s.typeCache, typeCacheEntry{typePtr: typPtr, info: info})
		}
		// If cache is full, don't cache (simple approach)
	}

	numFields := len(info.fields)
	if numFields == 0 {
		return Err("struct has no fields")
	}

	// Collect columns for SQL building
	var columns [32]string
	var colCount int

	for i := 0; i < numFields; i++ {
		fieldName := info.fields[i].Name
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
