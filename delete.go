package structsql

import (
	"unsafe"

	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

func (s *Structsql) Delete(structTable any, sql *string, values *[]any) error {
	if structTable == nil {
		return Err("no struct table provided")
	}

	v := structTable

	typ := tinyreflect.TypeOf(v)
	if typ.Kind() != K.Struct {
		return Err("input is not a struct")
	}

	if typ.Name() == "struct" {
		return Err("struct does not implement StructNamer interface")
	}

	// Use instance Conv
	c := s.convPool
	c.ResetBuffer(BuffOut)
	c.ResetBuffer(BuffWork)
	c.ResetBuffer(BuffErr)

	// Table name
	tableName := typ.Name()
	c.WrString(BuffOut, tableName)
	c.ToLower()
	tableStr := c.GetString(BuffOut)
	c.ResetBuffer(BuffOut)

	// Get cached type info
	typPtr := uintptr(unsafe.Pointer(typ))
	var info *typeInfo

	for _, entry := range s.typeCache {
		if entry.typePtr == typPtr {
			info = entry.info
			break
		}
	}

	if info == nil {
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
			s.convPool.WrString(BuffOut, field.Name.Name())
			s.convPool.ToLower()
			name := s.convPool.GetString(BuffOut)
			s.convPool.ResetBuffer(BuffOut)
			fields[i] = fieldInfo{Name: name}
		}
		info = &typeInfo{fields: fields}

		if len(s.typeCache) < cap(s.typeCache) {
			s.typeCache = append(s.typeCache, typeCacheEntry{typePtr: typPtr, info: info})
		}
	}

	// Find ID field
	idIndex := -1
	for i, field := range info.fields {
		if field.Name == "id" {
			idIndex = i
			break
		}
	}
	if idIndex == -1 {
		return Err("struct must have an 'id' field for delete")
	}

	// Build SQL
	c.WrString(BuffOut, "DELETE FROM ")
	c.WrString(BuffOut, tableStr)
	c.WrString(BuffOut, " WHERE id=")
	s.dbType.placeholder(1, c)

	*sql = c.GetStringZeroCopy(BuffOut)

	// Populate values
	*values = (*values)[:0]
	val := tinyreflect.ValueOf(v)
	fieldVal, err := val.Field(idIndex)
	if err != nil {
		return err
	}
	iface, err := fieldVal.Interface()
	if err != nil {
		return err
	}
	*values = append(*values, iface)

	return nil
}
