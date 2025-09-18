package structsql

import . "github.com/cdvelop/tinystring"

// placeholderSQLite generates SQLite-style placeholders (?, ?, ...)
func placeholderSQLite(index int, conv *Conv) {
	conv.WrString(BuffOut, "?")
}
