package app

func isSymbol(b byte) bool {
	return (CharTable[b] & FlagAllowSymbols) != 0
}

func isWhiteSpace(b byte) bool {
	return (CharTable[b] & FlagAllowWhitespace) != 0
}

func isNumber(b byte) bool {
	return (CharTable[b] & FlagAllowNumbers) != 0
}

func isAlpha(b byte) bool {
	return (CharTable[b] & (FlagAllowUpperOnly & FlagAllowLowerOnly)) != 0
}

func isSymbolOrWhitespace(b byte) bool {
	return isSymbol(b) || isWhiteSpace(b)
}

func (c *Config) DefaultMaskingByte () {
	if (c.UniversalFlags & FlagAllowSymbols) == 0 { 
		c.MaskingByte = 'x'

	} else {
		c.MaskingByte = '*'
	}
}

func NewEngine(Keyword []string) *Engine {
	trie := NewTrie()
	for _, k := range Keyword {
		trie.Insert([]byte(k))
	}

	return &Engine{
		Trie: trie,
	}
}

func (c *Config) MaskKeyword (buffer []byte, e *Engine ) []byte {
	if c.UniversalFlags&FlagAllowKeywords != 0 && e != nil && e.Trie != nil {
		e.Trie.SanitizeOrMask(buffer,c.MaskingByte, c.Cleaned)
	}

	return buffer
}