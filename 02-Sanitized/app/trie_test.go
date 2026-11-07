
package app

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
)

// ============================================================
// TEST FIXTURES
// ============================================================

func newTestFilter() (*Trie, *WhitelistTrie) {
	trie := NewTrie()
	whitelist := NewWhitelistTrie()

	keywords := []string{
		"hell",
		"damn",
		"shit",
		"fuck",
		"bitch",
		"crap",
	}

	allowedWords := []string{
		"hello",
	}

	for _, word := range keywords {
		trie.Insert([]byte(word))
	}

	for _, word := range allowedWords {
		whitelist.Insert([]byte(word))
	}

	return trie, whitelist
}

func assertOutput(
	t *testing.T,
	trie *Trie,
	whitelist *WhitelistTrie,
	input, expected string,
	remove bool,
) {
	t.Helper()

	buffer := []byte(input)

	result := trie.SanitizeOrMaskWithWhitelist(
		buffer,
		'*',
		remove,
		whitelist,
	)

	if string(result) != expected {
		t.Errorf(
			"input=%q got=%q want=%q",
			input,
			result,
			expected,
		)
	}
}

// ============================================================
// LEVEL 1: STANDARD DETECTION
// ============================================================

func TestLevel1_StandardDetection(t *testing.T) {
	trie, whitelist := newTestFilter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"canonical_word", "fuck", "****"},
		{"uppercase", "SHIT", "****"},
		{"mixed_case", "DaMn", "****"},
		{"multiple_matches", "fuck and shit", "**** and ****"},
		{"clean_input", "good morning", "good morning"},
		{"legitimate_word", "hello world", "hello world"},
		{"whitelist_with_punctuation", "hello!", "hello!"},
		{"toxic_after_whitelist", "hello hell", "hello ****"},
		{"suffix_policy", "hellish", "****ish"},
		{"numeric_boundary", "fuck123", "****123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(
				t,
				trie,
				whitelist,
				tt.input,
				tt.expected,
				false,
			)
		})
	}
}

// ============================================================
// LEVEL 2: OBFUSCATION DETECTION
// ============================================================

func TestLevel2_ObfuscationDetection(t *testing.T) {
	trie, whitelist := newTestFilter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"whitespace_injection", "f u c k", "* * * *"},
		{"dot_injection", "s.h.i.t", "*.*.*.*"},
		{"punctuation_injection", "f!u!c!k", "*!*!*!*"},
		{"repeated_characters", "fuuuck", "******"},
		{"uppercase_repetition", "SHIIIIT", "*******"},
		{"mixed_obfuscation", "f.u.u.c.k", "*.*.*.*.*"},
		{"whitelist_obfuscation", "h.e.l.l.o", "*.*.*.*.o"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(
				t,
				trie,
				whitelist,
				tt.input,
				tt.expected,
				false,
			)
		})
	}
}

// ============================================================
// LEVEL 3: ADVERSARIAL BOUNDARY
// ============================================================

func TestLevel3_AdversarialBoundary(t *testing.T) {
	trie, whitelist := newTestFilter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"embedded_prefix", "xhell", "xhell"},
		{"embedded_suffix", "hellx", "****x"},
		{"embedded_both_sides", "xhellx", "x****x"},
		{"long_embedded_word", "prefixhellsuffix", "prefix****suffix"},
		{"legitimate_prefix", "hello", "hello"},
		{"repeated_keywords", "hellhell", "********"},
		{"separator_heavy", "!!!!!!!!", "!!!!!!!!"},
		{"empty_input", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(
				t,
				trie,
				whitelist,
				tt.input,
				tt.expected,
				false,
			)
		})
	}
}

// ============================================================
// REMOVAL MODE
// ============================================================

func TestRemovalMode(t *testing.T) {
	trie, whitelist := newTestFilter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"canonical_removal", "hello fuck", "hello "},
		{"punctuation_preservation", "stop shit!", "stop !"},
		{"whitespace_injection", "f u c k", ""},
		{"whitelist_preservation", "hello world", "hello world"},
		{"multiple_removal", "fuck shit", ""},
		{"clean_input", "good morning", "good morning"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertOutput(
				t,
				trie,
				whitelist,
				tt.input,
				tt.expected,
				true,
			)
		})
	}
}

// ============================================================
// API COMPATIBILITY
// ============================================================

func TestLegacyAPI(t *testing.T) {
	trie, _ := newTestFilter()

	input := []byte("fuck")

	result := trie.SanitizeOrMask(
		input,
		'*',
		false,
	)

	if string(result) != "****" {
		t.Fatalf("unexpected result: %q", result)
	}
}

// ============================================================
// BUFFER INVARIANTS
// ============================================================

func TestBufferInvariants(t *testing.T) {
	trie, whitelist := newTestFilter()

	inputs := []string{
		"",
		"hello world",
		"fuck",
		"f u c k",
		"SHIT",
		"damn123",
		"hello hell",
	}

	for _, value := range inputs {
		t.Run(fmt.Sprintf("length_%d", len(value)), func(t *testing.T) {
			input := []byte(value)
			originalLength := len(input)

			result := trie.SanitizeOrMaskWithWhitelist(
				input,
				'*',
				false,
				whitelist,
			)

			if len(result) > originalLength {
				t.Fatalf(
					"output exceeded input length: got=%d max=%d",
					len(result),
					originalLength,
				)
			}
		})
	}
}

// ============================================================
// STRESS TESTING
// ============================================================

func TestStress_LargePayload(t *testing.T) {
	trie, whitelist := newTestFilter()

	sizes := []int{
		1024,
		10 * 1024,
		100 * 1024,
		1024 * 1024,
	}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("%d_bytes", size), func(t *testing.T) {
			input := bytes.Repeat([]byte("a"), size)

			result := trie.SanitizeOrMaskWithWhitelist(
				input,
				'*',
				false,
				whitelist,
			)

			if len(result) != size {
				t.Fatalf(
					"unexpected output length: got=%d want=%d",
					len(result),
					size,
				)
			}
		})
	}
}

func TestStress_SeparatorHeavy(t *testing.T) {
	trie, whitelist := newTestFilter()

	for _, size := range []int{64, 128, 256, 512, 1024, 2048} {
		t.Run(fmt.Sprintf("%d_bytes", size), func(t *testing.T) {
			input := bytes.Repeat([]byte("!"), size)

			result := trie.SanitizeOrMaskWithWhitelist(
				input,
				'*',
				false,
				whitelist,
			)

			if len(result) != size {
				t.Fatalf("unexpected output length")
			}
		})
	}
}

func TestStress_KeywordAtEnd(t *testing.T) {
	trie, whitelist := newTestFilter()

	input := append(
		bytes.Repeat([]byte("a"), 64*1024),
		[]byte(" fuck")...,
	)

	result := trie.SanitizeOrMaskWithWhitelist(
		input,
		'*',
		false,
		whitelist,
	)

	if !bytes.HasSuffix(result, []byte(" ****")) {
		t.Fatal("keyword at end was not detected")
	}
}

// ============================================================
// CONCURRENCY TESTING
// ============================================================

func TestConcurrency_ReadOnlyTrie(t *testing.T) {
	trie, whitelist := newTestFilter()

	const workers = 32
	const iterations = 100

	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for i := 0; i < iterations; i++ {
				input := []byte("hello fuck world")

				result := trie.SanitizeOrMaskWithWhitelist(
					input,
					'*',
					false,
					whitelist,
				)

				if string(result) != "hello **** world" {
					t.Errorf("unexpected output: %q", result)
					return
				}
			}
		}()
	}

	wg.Wait()
}

// ============================================================
// FUZZ TESTING
// ============================================================

func FuzzSanitizeOrMaskWithWhitelist(f *testing.F) {
	seeds := [][]byte{
		{},
		[]byte("hello"),
		[]byte("fuck"),
		[]byte("shit"),
		[]byte("f u c k"),
		[]byte("s.h.i.t"),
		[]byte("xhellx"),
		[]byte("!!!!!!!!"),
		{0xff, 0xfe, 0x00},
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	trie, whitelist := newTestFilter()

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			t.Skip()
		}

		input := append([]byte(nil), data...)

		result := trie.SanitizeOrMaskWithWhitelist(
			input,
			'*',
			false,
			whitelist,
		)

		if len(result) > len(input) {
			t.Fatal("output length exceeded input length")
		}
	})
}

// ============================================================
// BENCHMARK: FILTERING
// ============================================================

func BenchmarkContentFilter(b *testing.B) {
	trie, whitelist := newTestFilter()

	payload1KB := bytes.Repeat([]byte("a"), 1024)
	payload1KB[499] = ' '
	copy(payload1KB[500:504], []byte("hell"))
	payload1KB[504] = ' '

	payload4KB := bytes.Repeat([]byte("a"), 4096)

	payloadSeparator := bytes.Repeat([]byte("!"), 1024)

	tests := []struct {
		name    string
		payload []byte
		remove  bool
	}{
		{"L1_Canonical", []byte("hello fuck world"), false},
		{"L1_Standard_1KB", payload1KB, false},
		{"L2_Obfuscated", []byte("f.u.c.k"), false},
		{"L3_Clean_4KB", payload4KB, false},
		{"L3_Separator_1KB", payloadSeparator, false},
		{"Removal_1KB", payload1KB, true},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			buffer := make([]byte, len(tt.payload))

			b.ReportAllocs()
			b.SetBytes(int64(len(tt.payload)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				copy(buffer, tt.payload)

				_ = trie.SanitizeOrMaskWithWhitelist(
					buffer,
					'*',
					tt.remove,
					whitelist,
				)
			}
		})
	}
}

// ============================================================
// BENCHMARK: TRIE CONSTRUCTION
// ============================================================

func BenchmarkTrieConstruction(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = newTestFilter()
	}
}

// ============================================================
// BENCHMARK: WORST-CASE SCANNING
// ============================================================

func BenchmarkWorstCaseSeparatorScan(b *testing.B) {
	sizes := []int{
		64,
		128,
		256,
		512,
		1024,
		2048,
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d_bytes", size), func(b *testing.B) {
			trie, whitelist := newTestFilter()

			payload := bytes.Repeat([]byte("!"), size)
			buffer := make([]byte, size)

			b.ReportAllocs()
			b.SetBytes(int64(size))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				copy(buffer, payload)

				_ = trie.SanitizeOrMaskWithWhitelist(
					buffer,
					'*',
					false,
					whitelist,
				)
			}
		})
	}
}
