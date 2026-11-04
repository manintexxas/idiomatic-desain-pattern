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

func (t *Trie) Masking(b []byte, maskByte byte) {
	length := len(b)
	index := 0

	for index < length {
		current := t.Root
		lastEndIndex := -1
		innerIndex := index
		
		var lastChar byte = 0

		for innerIndex < length {
			char := b[innerIndex]

			if char == ' '{
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
			innerIndex ++


			if current.IsEnd{
				lastEndIndex = innerIndex
			}
		}

		if lastEndIndex != -1 {

			actualStart := index

			for actualStart < lastEndIndex && isWordBoundary(b, actualStart) {
				actualStart++
			}

			start := (actualStart == 0 || isWordBoundary(b, actualStart - 1))
			end   := (lastEndIndex == length || isWordBoundary(b, lastEndIndex))

			if start && end {
				for key := actualStart; key < lastEndIndex; key++ {
					b[key] = maskByte
				}

				index = lastEndIndex
			} else {
				index++
			}
		} else {
			index++
		}
	}
}