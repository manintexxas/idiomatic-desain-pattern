package app

func init() {
    for i := range 256 {
        b := byte(i)

        switch {
        case b >= 48 && b <= 57:
            charTable[b] |= FlagAllowNumbers

        case b >= 65 && b <= 90 :
            charTable[b] |= FlagAllowUpperOnly

        case b >= 97 && b <= 122:
            charTable[b] |= FlagAllowLowerOnly

        case (b >= 33 && b <= 47),
            (b >= 58 && b <= 64),
            (b >= 91 && b <= 96),
            (b >= 123 && b <= 126):
            charTable[b] |= FlagAllowSymbols

        case (b >= 9 && b <= 13), b == 32:
            charTable[b] |= FlagAllowWhitespace

        default:
            charTable[b] = 0
        }
    }
}