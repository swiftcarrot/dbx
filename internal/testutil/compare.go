package testutil

import "regexp"

func FormatSQL(sql string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(sql, " ")
}
