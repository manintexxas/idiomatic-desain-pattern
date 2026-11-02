package app

import (
	"errors"
)

type Flag uint16

type Config struct {
	AllowedFlag    Flag
	MaxLength      int
	BlacklistBytes [][]byte
	MaskingByte    byte
}

var (
	ErrEmptyInput                  = errors.New("Input cannot be empty")
	ErrPayloadTooLarge             = errors.New("Payload size exceeds maximum limit")
	ErrMalformedUnicode            = errors.New("Malformed unicode byte structure")
	ErrHTMLTagNotAllowed           = errors.New("HTML tag not allowed")
	ErrColoringNotAllowed          = errors.New("Text coloring or invalid color code not allowed")
	ErrSymbolNotAllowed            = errors.New("Symbol not allowed")
	ErrEmojiNotAllowed             = errors.New("Emoji not allowed")
	ErrSpecialAlphabetNotAllowed   = errors.New("Special alphabet not allowed")
	ErrSpecialCharNotAllowed       = errors.New("Special char not allowed")
	ErrNumberNotAllowed            = errors.New("Number not allowed")
	ErrDisallowedKeyword           = errors.New("Text contains blocked keywords")
	ErrWhitespaceOnly              = errors.New("Input contains only whitespaces")
	ErrInvalidControlChar          = errors.New("Invalid control character detected")
	ErrSqlInjectionPattern         = errors.New("Potential SQL injection pattern detected")
)

const (
	HTML Flag = 1 << iota
	Color
	Symbol
	Emoji
	SpecialChar
	SpecialAlpha
	Number
	Whitespace
	ControlChar
	SqlInjection
	Keyword
	Raw
)