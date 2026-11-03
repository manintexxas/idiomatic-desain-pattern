package app

import (
	"fmt"
	"strings"
)

func (v *ViolationReport) HasViolations() bool {
	return v.TotalViolations > 0
}

func (v *ViolationReport) Error() string {
	var sb strings.Builder
	sb.WriteString("validation failed:")

	if v.LetterViolations > 0 {
		fmt.Fprintf(&sb, " %d letter;", v.LetterViolations)
	}
	if v.NumberViolations > 0 {
		fmt.Fprintf(&sb, " %d number;", v.NumberViolations)
	}
	if v.SymbolViolations > 0 {
		fmt.Fprintf(&sb, " %d symbol;", v.SymbolViolations)
	}
	if v.WhitespaceViolations > 0 {
		fmt.Fprintf(&sb, " %d whitespace;", v.WhitespaceViolations)
	}
	if v.EmojiViolations > 0 {
		fmt.Fprintf(&sb, " %d emoji;", v.EmojiViolations)
	}
	if v.HTMLViolations > 0 {
		fmt.Fprintf(&sb, " %d html tag;", v.HTMLViolations)
	}
	if v.ColoringViolations > 0 {
		fmt.Fprintf(&sb, " %d coloring;", v.ColoringViolations)
	}
	if v.SpecialCharViolations > 0 {
		fmt.Fprintf(&sb, " %d special char;", v.SpecialCharViolations)
	}
	if v.ControlCharViolations > 0 {
		fmt.Fprintf(&sb, " %d control char;", v.ControlCharViolations)
	}
	if v.KeywordViolations > 0 {
		fmt.Fprintf(&sb, " %d keyword;", v.KeywordViolations)
	}
	if v.SQLiViolations > 0 {
		fmt.Fprintf(&sb, " %d sqli pattern;", v.SQLiViolations)
	}
	if v.MalformedUnicodeCount > 0 {
		fmt.Fprintf(&sb, " %d malformed unicode;", v.MalformedUnicodeCount)
	}

	return sb.String()
}
