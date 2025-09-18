package structsql

import . "github.com/cdvelop/tinystring"

// placeholderPostgre generates PostgreSQL-style placeholders ($1, $2, ...)
func placeholderPostgre(index int, conv *Conv) {
	conv.WrString(BuffOut, "$")
	// Convert int to string digits (no standard library)
	if index < 10 {
		conv.WrString(BuffOut, string(rune('0'+index)))
	} else {
		conv.WrString(BuffOut, string(rune('0'+index/10)))
		conv.WrString(BuffOut, string(rune('0'+index%10)))
	}
}
