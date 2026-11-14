package intake

type (
	Flag  uint8
	Types uint8
)

var charTable [256]Flag

const (
	Uppercase	Flag = 1 << iota
	Lowercase
	Symbolchar
	Whitespace
	Number
)

const (
	Mime Types = 1 << iota
	File
)

const (
	MaxFileNameBytes int = 0xFF
	MaxMIMEbytes     int = 0x80
)

const (
	MaxASCIIChar     uint8 = 0x7F
	CharDot          uint8 = 0x2E
	CharHyphen       uint8 = 0x2D
	CharSlash        uint8 = 0x2F
	CharBackslash    uint8 = 0x5C
	CharUnderscore   uint8 = 0x5F
	CharColon        uint8 = 0x3A
	CharSemiColon    uint8 = 0x3B
	CharPlus         uint8 = 0x2B
	CharHash         uint8 = 0x23
	CharMod          uint8 = 0x25
	CharAnd          uint8 = 0x26
	CharQuote        uint8 = 0x27
	Space            uint8 = 0x20
)
