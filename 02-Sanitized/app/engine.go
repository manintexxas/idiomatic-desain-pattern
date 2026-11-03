package app

func init() {
    for i := 0; i < 256; i++ {
        b := byte(i)

        switch {
        case (b >= 0 && b <= 31) || b == 127:
            CharTable[b] = 0 

        case b >= '0' && b <= '9':
            CharTable[b] |= FlagAllowNumbers

        case b >= 'A' && b <= 'Z':
            CharTable[b] |= FlagAllowUpperOnly

        case b >= 'a' && b <= 'z':
            CharTable[b] |= FlagAllowLowerOnly

        default:
            CharTable[b] |= FlagAllowSymbols
        }
    }
}