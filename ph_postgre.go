package structsql

import . "github.com/cdvelop/tinystring"

// placeholderPostgre generates PostgreSQL-style placeholders ($1, $2, ...)
func placeholderPostgre(index int, conv *Conv) {
	conv.WrString(BuffOut, "$")
	// Use AnyToBuff for tested integer-to-string conversion
	conv.AnyToBuff(BuffOut, index)
}
