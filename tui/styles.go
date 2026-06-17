package tui

import "github.com/fatih/color"

func ProviderStyle() *color.Color {
	return color.New(color.FgCyan, color.Bold)
}

func PathStyle() *color.Color {
	return color.New(color.Faint, color.Italic)
}

func SuccessMark() string {
	return color.GreenString("✔")
}

func ErrorMark() string {
	return color.RedString("✖")
}

func LoadingMark() string {
	return color.YellowString("⏳")
}

func HeaderStyle() *color.Color {
	return color.New(color.FgBlue, color.Bold)
}

func DimStyle() *color.Color {
	return color.New(color.Faint)
}

func Bold(s string) string {
	return color.New(color.Bold).Sprint(s)
}

func Cyan(s string) string {
	return color.CyanString(s)
}

func Green(s string) string {
	return color.GreenString(s)
}

func Red(s string) string {
	return color.RedString(s)
}

func Yellow(s string) string {
	return color.YellowString(s)
}

func Dim(s string) string {
	return color.New(color.Faint).Sprint(s)
}
