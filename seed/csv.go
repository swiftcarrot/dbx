package seed

// CSVImportOptions contains options for importing CSV data
type CSVImportOptions struct {
	// Delimiter specifies the field delimiter (default: ",")
	Delimiter string
	// NullValue specifies the string that represents NULL values (default: "")
	NullValue string
	// Header indicates whether the CSV file includes a header row (default: true)
	Header bool
	// Quote specifies the quote character (default: double quote)
	Quote string
	// Escape specifies the escape character (default: backslash)
	Escape string
	// Encoding specifies the file encoding (default: "UTF8")
	Encoding string
	// Columns specifies the target column names (default: use header row if present)
	Columns []string
}
