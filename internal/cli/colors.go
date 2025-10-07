package cli

import "fmt"

// ANSI color codes for consistent styling across all CLI commands
const (
	// Reset all formatting
	Reset = "\033[0m"

	// Text colors
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Gray    = "\033[90m"

	// Text formatting
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Italic = "\033[3m"

	// Background colors
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// Predefined color combinations for consistency
var (
	// Headers and titles
	HeaderStyle = Cyan + Bold
	TitleStyle  = Magenta + Bold

	// Status messages
	SuccessStyle = Green + Bold
	ErrorStyle   = Red + Bold
	WarningStyle = Yellow + Bold
	InfoStyle    = Blue + Bold

	// Data display
	LabelStyle     = Cyan
	ValueStyle     = White + Bold
	DimStyle       = Dim
	HighlightStyle = BgRed + White + Bold

	// Numbers and counts
	CountStyle = Yellow + Bold

	// Secondary info
	SecondaryStyle = Blue
	MetaStyle      = Gray
)

// Helper functions for common formatting patterns
func FormatHeader(text string) string {
	return HeaderStyle + text + Reset
}

func FormatTitle(text string) string {
	return TitleStyle + text + Reset
}

func FormatSuccess(text string) string {
	return SuccessStyle + text + Reset
}

func FormatError(text string) string {
	return ErrorStyle + text + Reset
}

func FormatWarning(text string) string {
	return WarningStyle + text + Reset
}

func FormatInfo(text string) string {
	return InfoStyle + text + Reset
}

func FormatLabel(text string) string {
	return LabelStyle + text + Reset
}

func FormatValue(text string) string {
	return ValueStyle + text + Reset
}

func FormatCount(count int) string {
	return CountStyle + fmt.Sprintf("%d", count) + Reset
}

func FormatHighlight(text string) string {
	return HighlightStyle + text + Reset
}

func FormatDim(text string) string {
	return DimStyle + text + Reset
}

func FormatSecondary(text string) string {
	return SecondaryStyle + text + Reset
}

func FormatMeta(text string) string {
	return MetaStyle + text + Reset
}

// Format a label-value pair
func FormatLabelValue(label, value string) string {
	return LabelStyle + label + Reset + " " + ValueStyle + value + Reset
}

// Format a count with label
func FormatCountLabel(label string, count int) string {
	return LabelStyle + label + Reset + " " + CountStyle + fmt.Sprintf("%d", count) + Reset
}
