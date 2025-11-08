// package styles define a color palette from nord & iceberg.vim
package style

import (
	"fmt"
	"image/color"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	lg "github.com/charmbracelet/lipgloss/v2"
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

func Headlinef(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	v := StyleHeadline.Render(s)
	fmt.Println(v)
}

func Added(s string) {
	v := StyleAdded.Render(s)
	fmt.Println(v)
}

func Addedf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	v := StyleAdded.Render(s)
	fmt.Println(v)
}

func Successf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	v := StyleAdded.Render(s)
	fmt.Println(v)
}

func Warningf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	v := StyleSecurity.Render(s)
	fmt.Println(v)
}

func Newline() { fmt.Println() }

func Fixed(s string) {
	v := StyleFixed.Render(s)
	fmt.Println(v)
}

func Styled(st lipgloss.Style) func(s string, a ...any) {
	return func(s string, a ...any) { fmt.Printf(s, a...) }
}

func Styledln(st lipgloss.Style) func(s string, a ...any) {
	return func(s string, a ...any) { fmt.Println(fmt.Sprintf(s, a...)) }
}

// Println wraps [fmt.Println] & [fmt.Sprintf]
func Println(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(msg)
}

var darkTheme = fang.ColorScheme{
	Base:           color.RGBA{25, 28, 35, 255},
	Title:          color.RGBA{129, 161, 193, 255},
	Description:    color.RGBA{180, 198, 211, 255},
	Codeblock:      color.RGBA{46, 52, 64, 255},
	Program:        color.RGBA{94, 129, 172, 255},
	DimmedArgument: color.RGBA{110, 115, 125, 255},
	Comment:        color.RGBA{76, 86, 106, 255},
	Flag:           color.RGBA{143, 188, 187, 255},
	FlagDefault:    color.RGBA{163, 190, 140, 255},
	Command:        color.RGBA{208, 135, 112, 255},
	QuotedString:   color.RGBA{136, 192, 208, 255},
	Argument:       color.RGBA{191, 97, 106, 255},
	Help:           color.RGBA{143, 188, 187, 255},
	Dash:           color.RGBA{216, 222, 233, 255},
	ErrorHeader: [2]color.Color{
		color.RGBA{236, 239, 244, 255},
		color.RGBA{191, 97, 106, 255},
	},
	ErrorDetails: color.RGBA{255, 203, 107, 255},
}

var lightTheme = fang.ColorScheme{
	Base:           color.RGBA{245, 247, 250, 255},
	Title:          color.RGBA{52, 73, 94, 255},
	Description:    color.RGBA{88, 110, 117, 255},
	Codeblock:      color.RGBA{230, 235, 240, 255},
	Program:        color.RGBA{70, 106, 145, 255},
	DimmedArgument: color.RGBA{140, 145, 155, 255},
	Comment:        color.RGBA{150, 160, 170, 255},
	Flag:           color.RGBA{0, 114, 178, 255},
	FlagDefault:    color.RGBA{106, 153, 85, 255},
	Command:        color.RGBA{217, 95, 2, 255},
	QuotedString:   color.RGBA{38, 139, 210, 255},
	Argument:       color.RGBA{203, 75, 22, 255},
	Help:           color.RGBA{0, 114, 178, 255},
	Dash:           color.RGBA{120, 130, 140, 255},
	ErrorHeader: [2]color.Color{
		color.RGBA{255, 255, 255, 255},
		color.RGBA{203, 75, 22, 255},
	},
	ErrorDetails: color.RGBA{230, 150, 50, 255},
}

func NewColorScheme(c lg.LightDarkFunc) fang.ColorScheme {
	return fang.ColorScheme{
		Base:           c(lightTheme.Base, darkTheme.Base),
		Title:          c(lightTheme.Title, darkTheme.Title),
		Description:    c(lightTheme.Description, darkTheme.Description),
		Codeblock:      c(lightTheme.Codeblock, darkTheme.Codeblock),
		Program:        c(lightTheme.Program, darkTheme.Program),
		DimmedArgument: c(lightTheme.DimmedArgument, darkTheme.DimmedArgument),
		Comment:        c(lightTheme.Comment, darkTheme.Comment),
		Flag:           c(lightTheme.Flag, darkTheme.Flag),
		FlagDefault:    c(lightTheme.FlagDefault, darkTheme.FlagDefault),
		Command:        c(lightTheme.Command, darkTheme.Command),
		QuotedString:   c(lightTheme.QuotedString, darkTheme.QuotedString),
		Argument:       c(lightTheme.Argument, darkTheme.Argument),
		Help:           c(lightTheme.Help, darkTheme.Help),
		Dash:           c(lightTheme.Dash, darkTheme.Dash),
		ErrorHeader: [2]color.Color{
			c(lightTheme.ErrorHeader[0], darkTheme.ErrorHeader[0]),
			c(lightTheme.ErrorHeader[1], darkTheme.ErrorHeader[1]),
		},
		ErrorDetails: c(lightTheme.ErrorDetails, darkTheme.ErrorDetails),
	}
}
