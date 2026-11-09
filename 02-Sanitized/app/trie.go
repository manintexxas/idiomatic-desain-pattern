package app

type TrieNode struct {
	Children [256]*TrieNode
	IsEnd    bool
}

type Trie struct {
	Root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{
		Root: &TrieNode{},
	}
}

func (t *Trie) Insert(b []byte) {
	current := t.Root
	for _, char := range b {
		if char >= 65 && char <= 90 {
			char |= 0x20
		}

		if current.Children[char] == nil {
			current.Children[char] = &TrieNode{}
		}
		current = current.Children[char]
	}
	current.IsEnd = true
}

type WhitelistTrie struct {
	Root *TrieNode
}

func NewWhitelistTrie() *WhitelistTrie {
	return &WhitelistTrie{
		Root: &TrieNode{},
	}
}

func (w *WhitelistTrie) Insert(b []byte) {
	current := w.Root
	for _, char := range b {
		if char >= 65 && char <= 90 {
			char |= 0x20
		}
		if current.Children[char] == nil {
			current.Children[char] = &TrieNode{}
		}
		current = current.Children[char]
	}
	current.IsEnd = true
}

func (w *WhitelistTrie) IsWhitelisted(b []byte, start, end int) bool {
	current := w.Root
	for i := start; i < end; i++ {
		char := b[i]
		if char >= 65 && char <= 90 {
			char |= 0x20
		}
		if current.Children[char] == nil {
			return false
		}
		current = current.Children[char]
	}
	return current.IsEnd
}

func isWordBoundary(b []byte, index int) bool {
	if index < 0 || index >= len(b) {
		return true
	}

	char := b[index]
	return isSymbolOrWhitespace(char) || isNumber(char)
}

func (t *Trie) findMatch(b []byte, index, length int) (int, int) {
	current := t.Root
	lastEndIndex := -1
	innerIndex := index
	var lastChar byte = 0

	for innerIndex < length {
		searchChar := b[innerIndex]

		if searchChar >= 65 && searchChar <= 90 {
			searchChar |= 0x20
		}

		next := current.Children[searchChar]

		if isSymbolOrWhitespace(searchChar) || isNumber(searchChar) || (searchChar == lastChar && next == nil) {
			innerIndex++
			continue
		}

		if next == nil {
			break
		}

		current = next
		lastChar = searchChar
		innerIndex++

		if current.IsEnd {
			lastEndIndex = innerIndex
		}
	}

	return lastEndIndex, innerIndex
}

func (t *Trie) hasMatch(b []byte) bool {
	length := len(b)
	if length == 0 || t.Root == nil {
		return false
	}

	for i := 0; i < length; i++ {
		current := t.Root
		innerIndex := i
		var lastChar byte = 0

		for innerIndex < length {
			searchChar := b[innerIndex]

			if searchChar >= 65 && searchChar <= 90 {
				searchChar |= 0x20
			}

			next := current.Children[searchChar]

			if isSymbolOrWhitespace(searchChar) || isNumber(searchChar) || (searchChar == lastChar && next == nil) {
				innerIndex++
				continue
			}

			if next == nil {
				break
			}

			current = next
			lastChar = searchChar
			innerIndex++

			if current.IsEnd {
				return true
			}
		}
	}

	return false
}

func (t *Trie) resolveActualStart(b []byte, index, lastEndIndex int) int {
	actualStart := index
	for actualStart < lastEndIndex && isWordBoundary(b, actualStart) {
		actualStart++
	}
	return actualStart
}

func (t *Trie) isValidBoundary(b []byte, actualStart, lastEndIndex, length int, matchLength int) bool {
	start := actualStart == 0 ||
		isWordBoundary(b, actualStart-1)

	isAtEnd := lastEndIndex == length

	var isNextLetter bool
	if !isAtEnd {
		nextChar := b[lastEndIndex]
		isNextLetter = (nextChar >= 'a' && nextChar <= 'z') ||
			(nextChar >= 'A' && nextChar <= 'Z')
	}

	if matchLength <= 3 {
		if isNextLetter {
			return false
		}

		if isAtEnd {
			return start
		}

		return start && isWordBoundary(b, lastEndIndex)
	}

	return true
}

func (t *Trie) applyMasking(b []byte, start, end int, maskByte byte) {
	for key := start; key < end; key++ {

		if isSymbolOrWhitespace(b[key]) {
			continue
		}

		b[key] = maskByte
	}
}

func (t *Trie) applyRemoval(b []byte, start, end, length int) (int, int) {
	n := copy(b[start:], b[end:length])
	newLength := start + n
	return newLength, start
}

func (t *Trie) SanitizeOrMaskWithWhitelist(b []byte, maskByte byte, isCleaned bool, whitelist *WhitelistTrie) []byte {
	length := len(b)
	index := 0

	for index < length {
		lastEndIndex, innerIndex := t.findMatch(b, index, length)

		if lastEndIndex != -1 {
			actualStart := t.resolveActualStart(b, index, lastEndIndex)
			matchLength := lastEndIndex - actualStart

			if t.isValidBoundary(b, actualStart, lastEndIndex, length, matchLength) {
				isWhitelisted := false
				fullWordEnd := actualStart
				if whitelist != nil {
					for fullWordEnd < length && ((b[fullWordEnd] >= 'a' && b[fullWordEnd] <= 'z') || (b[fullWordEnd] >= 'A' && b[fullWordEnd] <= 'Z')) {
						fullWordEnd++
					}
					if whitelist.IsWhitelisted(b, actualStart, fullWordEnd) {
						isWhitelisted = true
					}
				}

				if isWhitelisted {
					index = fullWordEnd
				} else {
					if isCleaned {
						if lastEndIndex < length && b[lastEndIndex] == ' ' {
							length, index = t.applyRemoval(b, actualStart, lastEndIndex+1, length)
						} else {
							length, index = t.applyRemoval(b, actualStart, lastEndIndex, length)
						}
					} else {
						t.applyMasking(b, actualStart, lastEndIndex, maskByte)
						index = lastEndIndex
					}
				}
			} else {
				index++
			}
		} else {
			if innerIndex > index {
				index = innerIndex
			} else {
				index++
			}
		}
	}

	return b[:length]
}

func (t *Trie) SanitizeOrMask(b []byte, maskByte byte, isCleaned bool) []byte {
	return t.SanitizeOrMaskWithWhitelist(b, maskByte, isCleaned, nil)
}
