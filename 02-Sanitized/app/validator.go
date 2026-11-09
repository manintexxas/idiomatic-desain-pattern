package app

func isSymbol(b byte) bool {
	return (charTable[b] & FlagAllowSymbols) != 0
}

func isWhiteSpace(b byte) bool {
	return (charTable[b] & FlagAllowWhitespace) != 0
}

func isNumber(b byte) bool {
	return (charTable[b] & FlagAllowNumbers) != 0
}

func isAlpha(b byte) bool {
	return (charTable[b] & (FlagAllowUpperOnly | FlagAllowLowerOnly)) != 0
}

func isSymbolOrWhitespace(b byte) bool {
	return isSymbol(b) || isWhiteSpace(b)
}

func countUnwhitelistedKeywordMatches(
	b []byte,
	trie *Trie,
	whitelist *WhitelistTrie,
) int {
	if trie == nil || trie.Root == nil || len(b) == 0 {
		return 0
	}

	count := 0
	length := len(b)

	for index := 0; index < length; {
		lastEndIndex, _ := trie.findMatch(
			b,
			index,
			length,
		)

		if lastEndIndex == -1 {
			index++
			continue
		}

		actualStart := trie.resolveActualStart(
			b,
			index,
			lastEndIndex,
		)

		matchLength := lastEndIndex - actualStart

		if !trie.isValidBoundary(
			b,
			actualStart,
			lastEndIndex,
			length,
			matchLength,
		) {
			index++
			continue
		}

		if whitelist != nil {
			fullWordEnd := actualStart

			for fullWordEnd < length {
				char := b[fullWordEnd]

				if !((char >= 'a' && char <= 'z') ||
					(char >= 'A' && char <= 'Z')) {
					break
				}

				fullWordEnd++
			}

			if fullWordEnd > actualStart &&
				whitelist.IsWhitelisted(
					b,
					actualStart,
					fullWordEnd,
				) {
				index = fullWordEnd
				continue
			}
		}

		count++
		index = lastEndIndex
	}

	return count
}

func NewEngine(BanKeyword []string, Keyword []string) *Engine {
	trie := NewTrie()
	for _, k := range BanKeyword {
		trie.Insert([]byte(k))
	}

	wltrie := NewWhitelistTrie()
	for _, k := range Keyword {
		trie.Insert([]byte(k))
	}

	return &Engine{
		Trie:   trie,
		WlTrie: wltrie,
	}
}

func (c *Config) MaskKeyword(buffer []byte, e *Engine) []byte {
	if e == nil || e.Trie == nil {
		return buffer
	}

	if (c.EngineModes & ModeFind) != 0 {
		_ = e.Trie.hasMatch(buffer)
	}

	if (c.EngineModes & ModeMask) != 0 {
		if e.WlTrie == nil {
			e.Trie.SanitizeOrMask(buffer, c.MaskingByte, false)
		} else {
			e.Trie.SanitizeOrMaskWithWhitelist(buffer, c.MaskingByte, false, e.WlTrie)
		}
	}

	if (c.EngineModes & ModeClean) != 0 {
		if e.WlTrie == nil {
			e.Trie.SanitizeOrMask(buffer, c.MaskingByte, true)
		} else {
			e.Trie.SanitizeOrMaskWithWhitelist(buffer, c.MaskingByte, true, e.WlTrie)
		}
	}

	return buffer
}
