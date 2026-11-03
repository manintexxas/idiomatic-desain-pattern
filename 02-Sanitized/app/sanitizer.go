package app

import (
	"errors"
)

type Flag uint16

const (
	FlagAllowLetters Flag = 1 << iota
	FlagAllowNumbers
	FlagAllowWhitespace
	FlagAllowSymbols
	FlagEnableAdvanced
)

const (
	FlagAllowMultibyte Flag = 1 << (iota + 5)
	FlagAllowHTML
	FlagAllowColoring
	FlagAllowEmoji
	FlagAllowSpecialChar
	FlagAllowControlChar
	FlagAllowKeywords
	FlagDetectSQLi
)

const (
	DefaultPreset  Flag = FlagAllowLetters
	AdvancedPreset Flag = FlagAllowMultibyte | FlagAllowHTML | FlagAllowColoring | FlagDetectSQLi
)

type Config struct {
	UniversalFlags Flag
	AdvancedFlags  Flag
	MaxLength      int
	MaskingByte    byte
	BlacklistBytes [][]byte
}

type Engine struct {
	Config    Config
	CharTable [256]uint16
}

type ViolationReport struct {
	LetterViolations      int
	NumberViolations      int
	SymbolViolations      int
	WhitespaceViolations  int
	EmojiViolations       int
	HTMLViolations        int
	ColoringViolations    int
	SpecialCharViolations int
	ControlCharViolations int
	KeywordViolations     int
	SQLiViolations        int
	MalformedUnicodeCount int
	TotalViolations       int
}

var (
	ErrEmptyInput                = errors.New("Input cannot be empty")
	ErrPayloadTooLarge           = errors.New("Payload size exceeds maximum limit")
	ErrMalformedUnicode          = errors.New("Malformed unicode byte structure")
	ErrHTMLTagNotAllowed         = errors.New("HTML tag not allowed")
	ErrColoringNotAllowed        = errors.New("Text coloring or invalid color code not allowed")
	ErrSymbolNotAllowed          = errors.New("Symbol not allowed")
	ErrEmojiNotAllowed           = errors.New("Emoji not allowed")
	ErrSpecialAlphabetNotAllowed = errors.New("Special alphabet not allowed")
	ErrSpecialCharNotAllowed     = errors.New("Special char not allowed")
	ErrNumberNotAllowed          = errors.New("Number not allowed")
	ErrDisallowedKeyword         = errors.New("Text contains blocked keywords")
	ErrWhitespaceOnly            = errors.New("Input contains only whitespaces")
	ErrInvalidControlChar        = errors.New("Invalid control character detected")
	ErrSqlInjectionPattern       = errors.New("Potential SQL injection pattern detected")
	ErrValidationFailed          = errors.New("Validation failed")
)