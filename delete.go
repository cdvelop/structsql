package structsql

import (
	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

func (s *Structsql) Delete(structTable any, sql *string, values *[]any) error {
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

	// Find ID field
	idIndex, err := s.findIdField(tableStr, info.fields, true)
	if err != nil {
		return err
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
