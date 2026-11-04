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
	return char == ' ' || char == '.' || char == ',' || char == '!' || char == '?' || char == '-'
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

		if char == lastChar && current.Children[char] == nil {
			innerIndex++
			continue
		}

		next := current.Children[char]
		if next == nil {
			break
		}

		current = next
		lastChar = char
		innerIndex++

		if current.IsEnd {
			lastEndIndex = innerIndex
		}
	}

	return lastEndIndex
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