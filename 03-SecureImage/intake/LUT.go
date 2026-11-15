package intake

type Flag uint8

var CharTable [256]Flag

const (
	Uppercase	Flag = 1 << iota
	Lowercase
	Symbolchar
	Whitespace
	Number
)

func init() {
	for i := range CharTable {
		b := byte(i)

		switch {
		case b >= '0' && b <= '9':
			CharTable[b] = Number

		case b >= 'A' && b <= 'Z':
			CharTable[b] = Uppercase

		case b >= 'a' && b <= 'z':
			CharTable[b] = Lowercase

		case (b >= 33 && b <= 47),
			(b >= 58 && b <= 64),
			(b >= 91 && b <= 96),
			(b >= 123 && b <= 126):
			CharTable[b] = Symbolchar

		case (b >= 9 && b <= 13), b == 32:
			CharTable[b] = Whitespace

		default:
			CharTable[b] = 0
		}
	}
}
