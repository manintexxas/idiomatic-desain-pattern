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
	LetterPreset Flag = FlagAllowUpperOnly | FlagAllowLowerOnly
	DefaultPreset  Flag = LetterPreset | FlagAllowNumbers | FlagAllowSymbols | FlagAllowWhitespace
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
	LetterViolations      		int
	MaxLetterViolations	  		int
	MinLetterViolations	  		int
	
	UpperViolations		  		int
	MaxUpperViolations	  		int
	MinUpperViolations	  		int
	
	LowerViolations		  		int
	MaxLowerViolations	 		int
	MinLowerViolations	  		int
	
	NumberViolations      		int
	MaxNumberViolations	  		int
	MinNumberViolations	  		int
	
	SymbolViolations      		int
	MaxSymbolViolations	  		int
	MinSymbolViolations	  		int
	
	WhitespaceViolations  		int
	MaxWhiteSpaceViolations		int
	MinWhiteSpaceViolations		int
	
	MinLengthViolations	  		int
	MaxLengthViolations	 		int
	
	TotalViolations       		int
	
	ReqOverMaxLetter	  		bool
	ReqOverMaxUpper		  		bool
	ReqOverMaxLower		  		bool
	ReqOverMaxNumber			bool
	ReqOverMaxSymbol	 		bool
	ReqOverMaxWhiteSpace		bool
	ReqOverMaxLength			bool

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
	ErrLengthLess				 = errors.New("Length less than minimal length")
	ErrLengthMore				 = errors.New("Length more than maximal length")
	ErrReqMoreMax				 = errors.New("Requirentment cannot surpass max lenght")
)