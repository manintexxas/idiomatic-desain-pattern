package app

import (
	"bytes"
	"testing"
)

func TestSanitizeOrMask(t *testing.T) {
	keywords := []string{"fuck", "shit", "badword"}
	trie := NewTrie()
	for _, kw := range keywords {
		trie.Insert([]byte(kw))
	}

	tests := []struct {
		name      string
		input     string
		maskByte  byte
		isCleaned bool
		expected  string
	}{
		{
			name:      "Exact Match Masking",
			input:     "there is a fuck here",
			maskByte:  '*',
			isCleaned: false,
			expected:  "there is a **** here",
		},
		{
			name:      "Exact Match Removal",
			input:     "there is a fuck here",
			maskByte:  '*',
			isCleaned: true,
			expected:  "there is a  here",
		},
		{
			name:      "Custom Mask Byte X",
			input:     "there is a fuck here",
			maskByte:  'x',
			isCleaned: false,
			expected:  "there is a xxxx here",
		},
		{
			name:      "Word Boundary False Positive Prevention",
			input:     "checking the fucking structure",
			maskByte:  '*',
			isCleaned: false,
			expected:  "checking the ****ing structure",
		},
		{
			name:      "Word Boundary False Positive Prevention Removal Mode",
			input:     "checking the fucking structure",
			maskByte:  '*',
			isCleaned: true,
			expected:  "checking the ing structure",
		},
		{
			name:      "Substring Affix Words",
			input:     "myfuckingshoes are dirty",
			maskByte:  '*',
			isCleaned: false,
			expected:  "my****ingshoes are dirty",
		},
		{
			name:      "Space Skipping Bypass Masking",
			input:     "what a f u c k you are",
			maskByte:  '*',
			isCleaned: false,
			expected:  "what a ******* you are",
		},
		{
			name:      "Space Skipping Bypass Removal",
			input:     "what a f u c k you are",
			maskByte:  '*',
			isCleaned: true,
			expected:  "what a  you are",
		},
		{
			name:      "Character Collapse Bypass Masking",
			input:     "you are fuuuuuckin crazy",
			maskByte:  '*',
			isCleaned: false,
			expected:  "you are ********in crazy",
		},
		{
			name:      "Character Collapse Bypass Removal",
			input:     "you are fuuuuuckin crazy",
			maskByte:  '*',
			isCleaned: true,
			expected:  "you are in crazy",
		},
		{
			name:      "Single Forbidden Word Input Masking",
			input:     "fuck",
			maskByte:  '*',
			isCleaned: false,
			expected:  "****",
		},
		{
			name:      "Single Forbidden Word Input Removal",
			input:     "fuck",
			maskByte:  '*',
			isCleaned: true,
			expected:  "",
		},
		{
			name:      "Input with Punctuation Masking",
			input:     "Watch out, fuck! Or shit?",
			maskByte:  '*',
			isCleaned: false,
			expected:  "Watch out, ****! Or ****?",
		},
		{
			name:      "Input with Punctuation Removal",
			input:     "Watch out, fuck! Or shit?",
			maskByte:  '*',
			isCleaned: true,
			expected:  "Watch out, ! Or ?",
		},
		{
			name:      "Clean Text Without Forbidden Words Masking",
			input:     "Good morning beautiful world",
			maskByte:  '*',
			isCleaned: false,
			expected:  "Good morning beautiful world",
		},
		{
			name:      "Clean Text Without Forbidden Words Removal",
			input:     "Good morning beautiful world",
			maskByte:  '*',
			isCleaned: true,
			expected:  "Good morning beautiful world",
		},
		{
			name:      "Multiple Occurrences Masking",
			input:     "shit in front shit in back",
			maskByte:  '*',
			isCleaned: false,
			expected:  "**** in front **** in back",
		},
		{
			name:      "Multiple Occurrences Removal",
			input:     "shit in front shit in back",
			maskByte:  '*',
			isCleaned: true,
			expected:  " in front  in back",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputBuffer := []byte(tt.input)
			result := trie.SanitizeOrMask(inputBuffer, tt.maskByte, tt.isCleaned)

			if !bytes.Equal(result, []byte(tt.expected)) {
				t.Errorf("\n[FAILED] %s\nInput    : %s\nGot      : %s\nExpected : %s",
					tt.name, tt.input, string(result), tt.expected)
			}
		})
	}
}

func BenchmarkSanitizeOrMask_MaskingMode(b *testing.B) {
	trie := NewTrie()
	keywords := []string{"fuck", "shit", "scam", "phishing", "hack"}
	for _, kw := range keywords {
		trie.Insert([]byte(kw))
	}

	input := []byte("Watch out there is a fuck and shit near the main hacking site!")
	maskByte := byte('*')

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := append([]byte(nil), input...)
		_ = trie.SanitizeOrMask(buf, maskByte, false)
	}
}

func BenchmarkSanitizeOrMask_RemovalMode(b *testing.B) {
	trie := NewTrie()
	keywords := []string{"fuck", "shit", "scam", "phishing", "hack"}
	for _, kw := range keywords {
		trie.Insert([]byte(kw))
	}

	input := []byte("Watch out there is a fuck and shit near the main hacking site!")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := append([]byte(nil), input...)
		_ = trie.SanitizeOrMask(buf, '*', true)
	}
}