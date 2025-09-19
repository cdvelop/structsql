package structsql

import (
	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

func (s *Structsql) Insert(structTable any, sql *string, values *[]any) error {
	typ, err := s.validateStruct(structTable)
	if err != nil {
		return err
	}

	// For now, handle only single struct (first one)
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
		s.dbType.placeholder(i+1, c)
	}

	c.WrString(BuffOut, ")")

	*sql = c.GetStringZeroCopy(BuffOut)

	// Populate values slice (reuse caller's buffer)
	*values = (*values)[:0] // Clear existing values

	// Ensure sufficient capacity
	if cap(*values) < numFields {
		// This should rarely happen in benchmarks, but handle gracefully
		*values = make([]any, 0, numFields)
	}

	val := tinyreflect.ValueOf(v)
	for i := 0; i < numFields; i++ {
		fieldVal, err := val.Field(i)
		if err != nil {
			return err
		}

		var iface any
		fieldVal.InterfaceZeroAlloc(&iface)

		*values = append(*values, iface) // Append to caller's buffer
	}

	return nil
}
