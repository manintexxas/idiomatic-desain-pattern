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
    ErrEmptyInput       		= errors.New("input cannot be empty")
    ErrPayloadTooLarge  		= errors.New("payload size exceeds maximum limit")
    ErrUpperNotAllowed  		= errors.New("uppercase characters are not allowed")
    ErrLowerNotAllowed  		= errors.New("lowercase characters are not allowed")
    ErrMalformedUnicode 		= errors.New("malformed unicode byte structure")
    ErrSymbolNotAllowed 		= errors.New("symbols are not allowed")
    ErrNumberNotAllowed 		= errors.New("numbers are not allowed")
    ErrDisallowedKeyword 		= errors.New("input contains blocked keywords")
    ErrWhitespaceOnly   		= errors.New("input contains only whitespace")
    ErrValidationFailed 		= errors.New("validation failed")

    ErrLengthLess 				= errors.New("length is below the minimum requirement")
    ErrLengthMore 				= errors.New("length exceeds the maximum requirement")
    ErrReqMoreMax 				= errors.New("requirement cannot exceed maximum length")

    ErrMinLetterViolation 		= errors.New("minimum letter requirement not met")
    ErrMaxLetterViolation 		= errors.New("maximum letter limit exceeded")

    ErrMinUpperViolation 		= errors.New("minimum uppercase requirement not met")
    ErrMaxUpperViolation 		= errors.New("maximum uppercase limit exceeded")

    ErrMinLowerViolation 		= errors.New("minimum lowercase requirement not met")
    ErrMaxLowerViolation 		= errors.New("maximum lowercase limit exceeded")

    ErrMinNumberViolation 		= errors.New("minimum number requirement not met")
    ErrMaxNumberViolation 		= errors.New("maximum number limit exceeded")

    ErrMinSymbolViolation 		= errors.New("minimum symbol requirement not met")
    ErrMaxSymbolViolation 		= errors.New("maximum symbol limit exceeded")

    ErrMinWhitespaceViolation 	= errors.New("minimum whitespace requirement not met")
    ErrMaxWhitespaceViolation 	= errors.New("maximum whitespace limit exceeded")
)
