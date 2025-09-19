package structsql

import (
	"unsafe"

	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

func (s *Structsql) Update(structTable any, sql *string, values *[]any) error {
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

	numFields := len(info.fields)
	if numFields == 0 {
		return Err("struct has no fields")
	}

	// Find primary key field index
	idIndex := -1
	for i, field := range info.fields {
		_, isPK := IDorPrimaryKey(tableStr, field.Name)
		if isPK {
			idIndex = i
			break
		}
	}
	if idIndex == -1 {
		return Err("struct must have a primary key field for update")
	}

	// Collect SET fields (all except id)
	var setColumns [32]string
	var setCount int
	for i := 0; i < numFields; i++ {
		if i != idIndex {
			setColumns[setCount] = info.fields[i].Name
			setCount++
		}
	}

	if setCount == 0 {
		return Err("no fields to update")
	}

	// Build SQL
	c.WrString(BuffOut, "UPDATE ")
	c.WrString(BuffOut, tableStr)
	c.WrString(BuffOut, " SET ")

	// SET clauses
	for i := 0; i < setCount; i++ {
		if i > 0 {
			c.WrString(BuffOut, ", ")
		}
		c.WrString(BuffOut, setColumns[i])
		c.WrString(BuffOut, "=")
		s.dbType.placeholder(i+1, c)
	}

	// WHERE
	c.WrString(BuffOut, " WHERE id=")
	s.dbType.placeholder(setCount+1, c)

	*sql = c.GetStringZeroCopy(BuffOut)

	// Populate values
	*values = (*values)[:0]
	val := tinyreflect.ValueOf(v)
	for i := 0; i < numFields; i++ {
		if i != idIndex {
			fieldVal, err := val.Field(i)
			if err != nil {
				return err
			}
			iface, err := fieldVal.Interface()
			if err != nil {
				return err
			}
			*values = append(*values, iface)
		}
	}
	// Add ID at the end
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
