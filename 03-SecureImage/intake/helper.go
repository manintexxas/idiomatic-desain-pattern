package intake

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

func shouldConvertToUnderscore(char byte) bool {
	isSym := (charTable[char] & Symbolchar) != 0
	isSpace := (charTable[char] & Whitespace) != 0
	
	allowedExceptions := char == CharDot || char == CharHyphen

	return (isSym || isSpace) && !allowedExceptions
}