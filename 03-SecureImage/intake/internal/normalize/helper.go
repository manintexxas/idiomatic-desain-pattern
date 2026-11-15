package normalize

import (
	in "03-SecureImage/intake"
	"strings"
)

func IsUpper(char byte) bool {
	return in.CharTable[char] == in.Uppercase
}

func IsLower(char byte) bool {
	return in.CharTable[char] == in.Lowercase
}

func IsSymbol(char byte) bool {
	return in.CharTable[char] == in.Symbolchar
}

func IsNumber(char byte) bool {
	return in.CharTable[char] == in.Number
}

func isWhiteSpace(char byte) bool {
	return in.CharTable[char] == in.Whitespace
}

func IsSymbolOrNonASCII(b byte) bool {
	return IsSymbol(b) || in.CharTable[b] == 0
}

func IsAllowedFileNameSymbol(b byte) bool {
	switch b {
	case CharHyphen, CharUnderscore, CharDot, CharLParen, CharRParen:
		return true
	}

	return false
}
func IsAllowedFileNameByte(b byte) bool {
	switch {
	case IsNumber(b), IsUpper(b), IsLower(b), IsAllowedFileNameSymbol(b):
		return true

	}
	return false
}

func IsAllowedMIMESymbol(b byte) bool {
	switch b {
	case CharPlus, CharHyphen, CharUnderscore, CharDot, CharHash, CharMod, CharAnd, CharQuote:
		return true
	}

	return false
}

func IsReservedWindowsDeviceName(baseName string) bool {
	if strings.EqualFold(baseName, "CON") || strings.EqualFold(baseName, "PRN") ||
		strings.EqualFold(baseName, "AUX") ||
		strings.EqualFold(baseName, "NUL") ||
		strings.EqualFold(baseName, "CLOCK$") ||
		strings.EqualFold(baseName, "CONIN$") ||
		strings.EqualFold(baseName, "CONOUT$") {
		return true
	}

	if len(baseName) == 4 {
		prefix := baseName[:3]
		digit := baseName[3]

		if digit >= '1' && digit <= '9' &&
			(strings.EqualFold(prefix, "COM") ||
				strings.EqualFold(prefix, "LPT")) {
			return true
		}
	}

	return false
}

func IsAllowedExt(s string) error {
	for i := range len(ext) {
		if s == ext[i] {
			return nil
		}
	}

	return in.ErrExtDoesntMatch
}

func containsNonASCIIAllowed(value string) error {
	for i := 0; i < len(value); i++ {
		if in.CharTable[value[i]] == 0 || value[i] > MaxASCIIChar {
			return in.ErrForbiddenCharacter
		}
	}

	return nil
}

func baseCharCheck(input string) error {
	if err := length(input); err != nil {
		return err
	}

	if err := containsNonASCIIAllowed(input); err != nil {
		return err
	}

	return nil
}
func ParseImageFormat(raw string) in.ImageFormat {
	switch raw {
	case string(in.FormatJPEG):
		return in.FormatJPEG

	case string(in.FormatJPG):
		return in.FormatJPG

	case string(in.FormatPNG):
		return in.FormatPNG

	case string(in.FormatWebP):
		return in.FormatWebP
	default:

		return in.FormatUnknown
	}
}

func ValidateMetadataConsistency(mime string, fileName string) ([2]in.ImageFormat, in.Status, error) {
	dotIndex := strings.LastIndexByte(fileName, CharDot)
	slashIndex := strings.LastIndexByte(mime, CharSlash)

	fileFmt := ParseImageFormat(fileName[dotIndex+1:])
	mimeFmt := ParseImageFormat(mime[slashIndex+1:])

	if fileFmt == in.FormatUnknown || mimeFmt == in.FormatUnknown {
		return [2]in.ImageFormat{}, in.REJECT, in.ErrMalformedMetadata
	}

	if fileFmt == mimeFmt || isSameFamily(fileFmt, mimeFmt) {
		return [2]in.ImageFormat{fileFmt}, in.PASS, nil
	}

	return [2]in.ImageFormat{fileFmt, mimeFmt}, in.FLAG, in.ErrMIMEMismatchFileName
}

func isSameFamily(a, b in.ImageFormat) bool {
	return (a == in.FormatJPG && b == in.FormatJPEG) || (a == in.FormatJPEG && b == in.FormatJPG)
}
