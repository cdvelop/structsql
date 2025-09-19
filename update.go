package structsql

import (
	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

func (s *Structsql) Update(structTable any, sql *string, values *[]any) error {
	typ, err := s.validateStruct(structTable)
	if err != nil {
		return err
	}

	v := structTable

	c := s.setupConv()

	var tableStr string
	s.getTableName(typ, &tableStr)

	info, err := s.getTypeInfo(typ)
	if err != nil {
		return err
	}

	numFields := len(info.fields)
	if numFields == 0 {
		return Err("struct has no fields")
	}

	// Find primary key field index
	idIndex, err := s.findIdField(tableStr, info.fields, true)
	if err != nil {
		return err
	}

	val := tinyreflect.ValueOf(v)

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

	// Populate values (only non-zero SET fields)
	*values = (*values)[:0]
	for i := 0; i < numFields; i++ {
		if i != idIndex {
			fieldVal, err := val.Field(i)
			if err != nil {
				return err
			}
			if !fieldVal.IsZero() {
				var iface any
				fieldVal.InterfaceZeroAlloc(&iface)
				*values = append(*values, iface)
			}
		}
	}
	// Add ID at the end
	fieldVal, err := val.Field(idIndex)
	if err != nil {
		return err
	}
	var iface any
	fieldVal.InterfaceZeroAlloc(&iface)
	*values = append(*values, iface)

	return nil
}
