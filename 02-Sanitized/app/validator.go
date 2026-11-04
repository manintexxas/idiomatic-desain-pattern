package app

// "unicode/utf8"

func (c *Config) MaskKeyword (buffer []byte, e *Engine ) []byte {

	if c.AdvancedFlags&FlagAllowKeywords != 0 && e != nil && e.Trie != nil {
		e.Trie.Masking(buffer, c.MaskingByte)
	}

	return buffer
}