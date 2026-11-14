package intake

import (
	"strings"
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
	return charTable[char] & Uppercase != 0
}

func IsLower(char byte) bool {
	return charTable[char] & Lowercase != 0
}

func IsSymbol(char byte) bool {
	return charTable[char] & Symbolchar != 0
}

func isNumber(char byte) bool {
	return charTable[char] & Number != 0
}

func isWhiteSpace(char byte) bool {
	return charTable[char] & Whitespace != 0
}

func isAllowedFileNameByte(char byte) bool {
	return IsLower(char) || 
			IsUpper(char) || 
			isNumber(char) || 
			char == CharHyphen || 
			char == CharUnderscore || 
			char == CharDot
}

func trimInput(s string) string{
	start := 0
	for start < len(s) && s[start] == Space {
		start++
	}
	end := len(s)
	for end > start && s[end-1] == Space {
		end--
	}

	if start == 0 && end == len(s) {
		return s
	}

	if start == end {
		return ""
	}

	return strings.Clone(s[start:end])
}

func toLower(s string) string {
	hasUpper := false
	for i := 0; i < len(s); i++ {
		if IsUpper(s[i]) {
			hasUpper = true
			break
		}
	}

	if !hasUpper {
		return s
	}

	buf := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if IsUpper(c) {
			c |= 32
		}
		buf[i] = c
	}
	
	return string(buf)
}

func lenCheck(name string) error {
	if name == "" {
		return ErrEmptyMetadata
	}

	return nil
}

func maxBytes(name string, t Types) error {
	length := len(name)

	if t == File {
		if length > MaxFileNameBytes {
			return ErrFileNameTooLong
		}
	}

	if t == Mime {
		if length > MaxMIMEbytes {
			return ErrMIMETooLong
		}
	}

	return nil
}

func isAllowedMIMESymbol(b byte) bool {
	switch b {
		case CharPlus, CharHyphen, CharUnderscore, CharDot, CharHash, CharMod, CharAnd, CharQuote:
			return  true
	}

	return  false
}

func isAllowedImageExtension(extension string) bool {
	return strings.EqualFold(extension, ".jpg") ||
		strings.EqualFold(extension, ".jpeg") ||
		strings.EqualFold(extension, ".png") ||
		strings.EqualFold(extension, ".webp")
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


func isReservedWindowsDeviceName(baseName string) bool {
	if strings.EqualFold(baseName, "CON") ||
		strings.EqualFold(baseName, "PRN") ||
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

func containsNonASCIIAllowed(value string) error {
	for i := 0; i < len(value); i++ {
		if charTable[value[i]] == 0 || value[i] > MaxASCIIChar {
			return ErrForbiddenCharacter
		}
	}

	return nil
}

func baseCharCheck(input string) error {
	if err := lenCheck(input); err != nil {
		return err
	}

	if err := containsNonASCIIAllowed(input); err != nil {
		return err
	}

	return nil
}

func validateMetadataConsistency( mime string, fileName string ) error {
	dotIndex := strings.LastIndexByte(fileName, CharDot)

	if dotIndex <= 0 || dotIndex == len(fileName)-1 {
		return ErrInvalidFileName
	}

	extension := fileName[dotIndex:]

	if !mimeMatchesExtension(mime, extension) {
		return ErrMIMEMismatchFileName
	}

	return nil
}