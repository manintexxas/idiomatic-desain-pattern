package app

type TrieNode struct {
	Node map[byte]*TrieNode
	IsEnd	 bool
}

type Trie struct {
	Root *TrieNode
}

func NewTrie() *Trie {
	return &Trie {
		Root: &TrieNode{
			Node: make(map[byte]*TrieNode),
		},
	}
}

func (t *Trie) Insert (b []byte) {
	current := t.Root
	for _, char := range b {
		if current.Node[char] == nil {
			current.Node[char] = &TrieNode{
				Node: make(map[byte]*TrieNode),
			}
		}
		current = current.Node[char]
	}
	current.IsEnd = true
}

func (t *Trie) Searchar (b []byte) bool {
	maxLength := len(b)

	for index := 0; index < maxLength; index++ {
		current := t.Root
		inner := index
		matched := false

		for inner < maxLength {
			char := b[inner]

			if next, ok := current.Node[char]; ok {
				current = next
				matched = true

				if current.IsEnd{
					return true
				}

				nextInner := inner + 1
				for nextInner < maxLength && (b[nextInner] == char || b[nextInner] == ' '){
					nextInner++
				}
				
				inner = nextInner

			} else {
				break
			}
		}

		if matched && current.IsEnd {
			return true
		}
	}

	return false
}