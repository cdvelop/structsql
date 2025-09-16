package structsql

import (
	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
)

func Insert(v any) (string, []interface{}, error) {
	// Check if implements StructNamer
	namer, ok := v.(StructNamer)
	if !ok {
		return "", nil, Err("struct does not implement StructNamer interface")
	}

	typ := tinyreflect.TypeOf(v)
	if typ.Kind() != 25 {
		return "", nil, Err("input is not a struct")
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

	// Get fields
	numFields, err := typ.NumField()
	if err != nil {
		return "", nil, err
	}

	if numFields == 0 {
		return "", nil, Err("struct has no fields")
	}

	// Collect columns and values
	var columns []string
	var values []interface{}
	var fieldName string

	for i := 0; i < numFields; i++ {
		field, err := typ.Field(i)
		if err != nil {
			return "", nil, err
		}

		c.WrString(BuffOut, field.Name.Name())
		c.ToLower()
		fieldName = c.GetString(BuffOut)
		c.ResetBuffer(BuffOut)

		columns = append(columns, fieldName)

		// Get value
		val := tinyreflect.ValueOf(v)
		fieldVal, err := val.Field(i)
		if err != nil {
			return "", nil, err
		}

		iface, err := fieldVal.Interface()
		if err != nil {
			return "", nil, err
		}

		values = append(values, iface)
	}

	// Build SQL
	c.WrString(BuffOut, "INSERT INTO ")
	c.WrString(BuffOut, tableStr)
	c.WrString(BuffOut, " (")

	// Columns
	for i, col := range columns {
		if i > 0 {
			c.WrString(BuffOut, ", ")
		}
		c.WrString(BuffOut, col)
	}

	c.WrString(BuffOut, ") VALUES (")

	// Placeholders
	for i := 0; i < len(columns); i++ {
		if i > 0 {
			c.WrString(BuffOut, ", ")
		}
		c.WrString(BuffOut, "?")
	}

	c.WrString(BuffOut, ")")

	sql := c.GetString(BuffOut)

	return sql, values, nil
}
