package structsql

import (
	"github.com/cdvelop/tinyreflect"
	. "github.com/cdvelop/tinystring"
	"unsafe"
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

	// Get cached type info
	typPtr := uintptr(unsafe.Pointer(typ))
	typeInfo, ok := typeCache[typPtr]
	if !ok {
		// Build cache
		numFields, err := typ.NumField()
		if err != nil {
			return "", nil, err
		}
		fields := make([]FieldInfo, numFields)
		for i := 0; i < numFields; i++ {
			field, err := typ.Field(i)
			if err != nil {
				return "", nil, err
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
		typeCache[typPtr] = typeInfo
	}

	numFields := len(typeInfo.fields)
	if numFields == 0 {
		return "", nil, Err("struct has no fields")
	}

	// Collect columns and values
	var columns [32]string
	var values [32]interface{}
	var colCount, valCount int

	for i := 0; i < numFields; i++ {
		fieldName := typeInfo.fields[i].Name

		columns[colCount] = fieldName
		colCount++

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

		values[valCount] = iface
		valCount++
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

	sql := c.GetString(BuffOut)

	return sql, values[:valCount], nil
}
