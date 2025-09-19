package structsql

import (
	"unsafe"

	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

func (s *Structsql) validateStruct(structTable any) (*tinyreflect.Type, error) {
	if structTable == nil {
		return nil, Err("no struct table provided")
	}

	v := structTable

	typ := tinyreflect.TypeOf(v)
	if typ.Kind() != K.Struct {
		return nil, Err("input is not a struct")
	}

	if typ.Name() == "struct" {
		return nil, Err("struct does not implement StructNamer interface")
	}

	return typ, nil
}

func (s *Structsql) setupConv() *Conv {
	c := s.convPool
	c.ResetBuffer(BuffOut)
	c.ResetBuffer(BuffWork)
	c.ResetBuffer(BuffErr)
	return c
}

func (s *Structsql) getTableName(typ *tinyreflect.Type, tableStr *string) {
	c := s.convPool
	tableName := typ.Name()
	c.WrString(BuffOut, tableName)
	c.ToLower()
	*tableStr = c.GetString(BuffOut)
	c.ResetBuffer(BuffOut)
}

func (s *Structsql) getTypeInfo(typ *tinyreflect.Type) (*typeInfo, error) {
	typPtr := uintptr(unsafe.Pointer(typ))
	var foundInfo *typeInfo

	for _, entry := range s.typeCache {
		if entry.typePtr == typPtr {
			foundInfo = entry.info
			break
		}
	}

	if foundInfo == nil {
		numFields, err := typ.NumField()
		if err != nil {
			return nil, err
		}
		fields := make([]fieldInfo, numFields)
		for i := 0; i < numFields; i++ {
			field, err := typ.Field(i)
			if err != nil {
				return nil, err
			}
			s.convPool.WrString(BuffOut, field.Name.Name())
			s.convPool.ToLower()
			name := s.convPool.GetString(BuffOut)
			s.convPool.ResetBuffer(BuffOut)
			fields[i] = fieldInfo{Name: name}
		}
		foundInfo = &typeInfo{fields: fields}

		if len(s.typeCache) < cap(s.typeCache) {
			s.typeCache = append(s.typeCache, typeCacheEntry{typePtr: typPtr, info: foundInfo})
		}
	}

	return foundInfo, nil
}

func (s *Structsql) findIdField(tableStr string, fields []fieldInfo, required bool) (int, error) {
	idIndex := -1
	for i, field := range fields {
		_, isPK := IDorPrimaryKey(tableStr, field.Name)
		if isPK {
			idIndex = i
			break
		}
	}

	if idIndex == -1 && required {
		return -1, Err("struct must have a primary key field")
	}

	return idIndex, nil
}
