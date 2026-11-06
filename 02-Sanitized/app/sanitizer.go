package app

import (
	"errors"
)

type Flag uint16

var (
	CharTable [256]Flag
)

const (
	FlagAllowUpperOnly 	  Flag = 1 << iota
	FlagAllowLowerOnly
	FlagAllowNumbers
	FlagAllowWhitespace
	FlagAllowSymbols
	FlagAllowKeywords
)

const (
	SentencePreset Flag = FlagAllowUpperOnly | FlagAllowLowerOnly
	DefaultPreset  Flag = SentencePreset | FlagAllowNumbers | FlagAllowSymbols | FlagAllowWhitespace
)

type Engine struct {
	Trie *Trie
}

type Config struct {
	UniversalFlags Flag
	MaxLength      int
	Cleaned		   bool
	MaskingByte    byte
}

type ViolationReport struct {
	UpperViolations		  int
	LowerViolations		  int
	LetterViolations      int
	NumberViolations      int
	SymbolViolations      int
	WhitespaceViolations  int
	TotalViolations       int
}

var (
	ErrEmptyInput                = errors.New("Input cannot be empty")
	ErrPayloadTooLarge           = errors.New("Payload size exceeds maximum limit")
	ErrUpperNotAllowed           = errors.New("Upper char not allowed")
	ErrLowerNotAllowed           = errors.New("Lower char not allowed")
	ErrMalformedUnicode          = errors.New("Malformed unicode byte structure")
	ErrSymbolNotAllowed          = errors.New("Symbol not allowed")
	ErrNumberNotAllowed          = errors.New("Number not allowed")
	ErrDisallowedKeyword         = errors.New("Text contains blocked keywords")
	ErrWhitespaceOnly            = errors.New("Input contains only whitespaces")
	ErrValidationFailed          = errors.New("Validation failed")
)