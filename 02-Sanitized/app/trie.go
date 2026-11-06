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

func isWordBoundary(b []byte, index int) bool {
	if index < 0 || index >= len(b) {
		return true
	}

	char := b[index]
	return (CharTable[char] & (FlagAllowSymbols | FlagAllowWhitespace)) != 0
}

func (t *Trie) findMatch(b []byte, index, length int) int {
	current := t.Root
	lastEndIndex := -1
	innerIndex := index
	var lastChar byte = 0

	for innerIndex < length {
		char := b[innerIndex]

		if char == ' ' {
			innerIndex++
			continue
		}

		searchChar := char
		if searchChar >= 65 && searchChar <= 90 {
			searchChar |= 0x20
		}

		if searchChar == lastChar && current.Children[searchChar] == nil {
			innerIndex++
			continue
		}

		next := current.Children[searchChar]
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

	return lastEndIndex
}

func (t *Trie) hasMatch (b []byte) bool {
	length := len(b)
	for index := 0;index < length;index++{
		if t.findMatch(b, index, length) != -1 {
			return true
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

func (t *Trie) isValidBoundary(b []byte, actualStart, lastEndIndex, length int) bool {
	start := (actualStart == 0 || isWordBoundary(b, actualStart-1))
	
	if lastEndIndex == length {
		return start
	}

	nextChar := b[lastEndIndex]
	isNextLetter := (nextChar >= 'a' && nextChar <= 'z') || (nextChar >= 'A' && nextChar <= 'Z')

	if isNextLetter {
		return true
	}

	end := isWordBoundary(b, lastEndIndex)
	return start && end
}

func (t *Trie) applyMasking(b []byte, start, end int, maskByte byte) {
	for key := start; key < end; key++ {
		b[key] = maskByte
	}
}

func (t *Trie) applyRemoval(b []byte, start, end, length int) (int, int) {
	n := copy(b[start:], b[end:length])
	newLength := start + n
	return newLength, start
}

func (t *Trie) SanitizeOrMask(b []byte, maskByte byte, isCleaned bool) []byte {
	length := len(b)
	index := 0

	for index < length {
		lastEndIndex := t.findMatch(b, index, length)

		if lastEndIndex != -1 {
			actualStart := t.resolveActualStart(b, index, lastEndIndex)

			if t.isValidBoundary(b, actualStart, lastEndIndex, length) {
				if isCleaned {
					length, index = t.applyRemoval(b, actualStart, lastEndIndex, length)
				} else {
					t.applyMasking(b, actualStart, lastEndIndex, maskByte)
					index = lastEndIndex
				}
			} else {
				index++
			}
		} else {
			index++
		}
	}

	return b[:length]
}