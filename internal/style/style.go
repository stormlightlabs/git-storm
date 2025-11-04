// package styles define a color palette from nord & iceberg.vim
package style

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	Background     = lipgloss.Color("#1B1F27") // dark slate blue/charcoal
	Foreground     = lipgloss.Color("#D8DEE9") // light grey-white for readable text
	AccentBlue     = lipgloss.Color("#71A6D2") // main iceberg blue
	AccentSteel    = lipgloss.Color("#4484B4") // steel blue
	AccentLapis    = lipgloss.Color("#1B5E98") // lapis lazuli
	AccentCerulean = lipgloss.Color("#074683") // dark cerulean
	AddedColor     = lipgloss.Color("#a3be8c") // nord aurora
	ChangedColor   = AccentSteel
	FixedColor     = AccentCerulean
	RemovedColor   = lipgloss.Color("#BF616A") // nord aurora
	SecurityColor  = lipgloss.Color("#d08770") // nord aurora
)

var (
	StyleHeadline = fgColor(AccentBlue).Bold(true)
	StyleText     = fgColor(Foreground).Background(Background)
	StyleAdded    = fgColor(AddedColor)
	StyleChanged  = fgColor(ChangedColor)
	StyleFixed    = fgColor(FixedColor)
	StyleRemoved  = fgColor(RemovedColor)
	StyleSecurity = fgColor(SecurityColor)
)

func fgColor(c lipgloss.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(c) }

func Headline(s string) {
	v := StyleHeadline.Render(s)
	fmt.Println(v)
}

func Added(s string) {
	v := StyleAdded.Render(s)
	fmt.Println(v)
}

func Fixed(s string) {
	v := StyleFixed.Render(s)
	fmt.Println(v)
}

func Styled(st lipgloss.Style) func(s string, a ...any) {
	return func(s string, a ...any) {
		fmt.Printf(s, a...)
	}
}

func Styledln(st lipgloss.Style) func(s string, a ...any) {
	return func(s string, a ...any) {
		fmt.Println(fmt.Sprintf(s, a...))
	}
}
