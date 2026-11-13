package intake

import "strings"

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

func cleanedInput(s string) string {
	start := 0
	for start < len(s) && isWhiteSpace(s[start]) {
		start++
	}
	end := len(s)
	for end > start && isWhiteSpace(s[end-1]) {
		end--
	}
	s = s[start:end]

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

func lenCheck(name string, length int) error {
	if name == "" {
		return ErrEmptyString
	}

	if len(name) > length {
		return ErrFileNameTooLong
	}

	return nil
}

func tokenCheck(b byte) bool {
	switch b {
		case CharPlus, CharHyphen, CharUnderscore, CharDot, CharHash, CharMod, CharAnd, CharQuote, CharSlash:
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

func mimeMatchesExtension(format ImageFormat, extension string) bool {
	switch format {
	case FormatJPEG:
		return strings.EqualFold(extension, ".jpg") ||
			strings.EqualFold(extension, ".jpeg")

	case FormatPNG:
		return strings.EqualFold(extension, ".png")

	case FormatWebP:
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

func containsNonASCIIByte(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] > MaxASCIIChar {
			return true
		}
	}

	return false
}