package shared

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var caser = cases.Title(language.English)

func TitleCase(s string) string {
	return caser.String(s)
}
