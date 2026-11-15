package normalize

import (
	"strings"
	in "03-SecureImage/intake"

)

func init() {
	for i := range charTable {
		b := byte(i)

		switch {
		case b >= '0' && b <= '9':
			charTable[b] = Number

		case b >= 'A' && b <= 'Z':
			charTable[b] = Uppercase

		case b >= 'a' && b <= 'z':
			charTable[b] = Lowercase

		case (b >= 33 && b <= 47),
			(b >= 58 && b <= 64),
			(b >= 91 && b <= 96),
			(b >= 123 && b <= 126):
			charTable[b] = Symbolchar

		case (b >= 9 && b <= 13), b == 32:
			charTable[b] = Whitespace

		default:
			charTable[b] = 0
		}
	}
}

func IsUpper(char byte) bool {
	return charTable[char] == Uppercase
}

func IsLower(char byte) bool {
	return charTable[char] == Lowercase
}

func IsSymbol(char byte) bool {
	return charTable[char] == Symbolchar
}

func IsNumber(char byte) bool {
	return charTable[char] == Number
}

func isWhiteSpace(char byte) bool {
	return charTable[char] == Whitespace
}

func IsSymbolOrNonASCII(b byte) bool {
	return IsSymbol(b) || charTable[b] == 0
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
			return  true
	}

	return  false
}

func IsReservedWindowsDeviceName(baseName string) bool {
	if strings.EqualFold(baseName, "CON") ||strings.EqualFold(baseName, "PRN") ||
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
		if s == ext[i]   {
			return nil
		}
	}

	return in.ErrExtDoesntMatch
}

func mimeMatchesExtension(mime string, extension string) bool {
	switch mime {
	case "image/jpeg":
		return strings.EqualFold(extension, ".jpg") ||
			strings.EqualFold(extension, ".jpeg")

	case "image/png":
		return strings.EqualFold(extension, ".png")

	case "image/webp":
		return strings.EqualFold(extension, ".webp")

	default:
		return false
	}
}

func containsNonASCIIAllowed(value string) error {
	for i := 0; i < len(value); i++ {
		if charTable[value[i]] == 0 || value[i] > MaxASCIIChar {
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

func ValidateMetadataConsistency( mime string, fileName string ) in.Status {
	dotIndex := strings.LastIndexByte(fileName, CharDot)

	extension := fileName[dotIndex:]

	if !mimeMatchesExtension(mime, extension) {
		return in.FLAG
	}

	return in.PASS
}