package app

type TrieNode struct {
	Node map[rune]*TrieNode
	End	 bool
}

type Trie struct {
	Root *TrieNode
}

func NewTrie() *Trie {
	return &Trie {
		Root: &TrieNode{
			Node: make(map[rune]*TrieNode),
		},
	}
}

func (t *Trie) Insert (word string) {
	current := t.Root
	for _, ch := range word {
		if current.Node[ch] == nil {
			current.Node[ch] = &TrieNode{
				Node: make(map[rune]*TrieNode),
			}
		}
		current = current.Node[ch]
	}
	current.End = true
}

func (t *Trie) Search (input string) bool {
	runes := []rune(input)
	n := len(runes)

	for i := 0; i < n; i++ {
		current := t.Root
		j := i
		matched := false

		for j < n {
			ch := runes[j]

			if next, ok := current.Node[ch]; ok {
				current = next
				matched = true

				if current.End{
					return true
				}

				nextJ := j + 1
				for nextJ < n && (runes[nextJ] == ch || runes[nextJ] == ' '){
					nextJ++
				}
				
				j = nextJ

			} else {
				break
			}
		}

		if matched && current.End {
			return true
		}
	}

	return false
}