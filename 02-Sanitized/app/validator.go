package app

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