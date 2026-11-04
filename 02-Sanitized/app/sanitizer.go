package app

import (
	"errors"

	"github.com/microcosm-cc/bluemonday"
)

type Flag uint16

var (
	CharTable [256]Flag
)

const (
	FlagAllowUpperOnly Flag = 1 << iota
	FlagAllowLowerOnly
	FlagAllowNumbers
	FlagAllowWhitespace
	FlagAllowSymbols
	FlagEnableAdvanced
)

const (
	FlagAllowKeywords Flag = 1 << (iota + 6)
	FlagAllowColoring
	FlagAllowMultibyte
	FlagAllowSpecialChar
	FlagAllowEmoji
	FlagAllowHTML
	FlagDetectSQLi
)

const (
	SentencePreset Flag = FlagAllowUpperOnly | FlagAllowLowerOnly
	DefaultPreset  Flag = SentencePreset | FlagAllowNumbers | FlagAllowSymbols | FlagAllowWhitespace
	AdvancedPreset Flag = FlagAllowMultibyte | FlagAllowHTML | FlagAllowColoring | FlagDetectSQLi
)

type Engine struct {
	Trie *Trie
}

func NewEngine(Keyword []string) *Engine {
	trie := NewTrie()
	for _, k := range Keyword {
		trie.Insert([]byte(k))
	}

	return &Engine{
		Trie: trie,
	}
}

type Config struct {
	UniversalFlags Flag
	AdvancedFlags  Flag
	MaxLength      int
	MaskingByte    byte
	HTMLConfig     *HTMLConfig
	SQLiConfig     *SQLiConfig
}

type HTMLConfig struct {
	Policy *bluemonday.Policy
}

type SQLiConfig struct {
	StrictnessLevel int
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
	ErrSqlInjectionPattern       = errors.New("Potential SQL injection pattern detected")
	ErrValidationFailed          = errors.New("Validation failed")
)