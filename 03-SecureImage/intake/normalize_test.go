
package intake

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// ============================================================
// LEVEL 1 — UNIT TESTS
// Input umum: kondisi normal dan penggunaan sehari-hari.
// ============================================================

func TestNormalizeMIME(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		// Valid image MIME.
		{
			name:  "jpeg",
			input: "image/jpeg",
			want:  "image/jpeg",
		},
		{
			name:  "png",
			input: "image/png",
			want:  "image/png",
		},
		{
			name:  "webp",
			input: "image/webp",
			want:  "image/webp",
		},

		// Case normalization.
		{
			name:  "uppercase jpeg",
			input: "IMAGE/JPEG",
			want:  "image/jpeg",
		},
		{
			name:  "mixed case png",
			input: "Image/PnG",
			want:  "image/png",
		},
		{
			name:  "mixed case webp",
			input: "ImAgE/WeBp",
			want:  "image/webp",
		},

		// Outer whitespace.
		{
			name:  "leading whitespace",
			input: "  image/jpeg",
			want:  "image/jpeg",
		},
		{
			name:  "trailing whitespace",
			input: "image/png  ",
			want:  "image/png",
		},
		{
			name:  "both outer whitespace",
			input: "  image/webp  ",
			want:  "image/webp",
		},

		{
			name:    "application pdf",
			input:   "application/pdf",
			wantErr: true,
		},
		{
			name:    "text plain",
			input:   "text/plain",
			wantErr: true,
		},
		{
			name:    "audio mime",
			input:   "audio/mpeg",
			wantErr: true,
		},
		{
			name:    "video mime",
			input:   "video/mp4",
			wantErr: true,
		},

		// Unsupported image MIME.
		{
			name:    "gif",
			input:   "image/gif",
			wantErr: true,
		},
		{
			name:    "svg",
			input:   "image/svg+xml",
			wantErr: true,
		},
		{
			name:    "tiff",
			input:   "image/tiff",
			wantErr: true,
		},
		{
			name:    "bmp",
			input:   "image/bmp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMIME(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if got != "" {
					t.Fatalf("expected empty result on error, got %q", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf(
					"NormalizeMIME() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

// ============================================================
// LEVEL 2 — ROBUSTNESS TESTS
// Input abnormal: malformed, boundary, dan kondisi tidak lazim.
// ============================================================

// ------------------------------------------------------------
// Malformed MIME
// ------------------------------------------------------------

func TestNormalizeMIMEInvalidSyntax(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty",
			input: "",
		},
		{
			name:  "whitespace only",
			input: "   ",
		},
		{
			name:  "missing subtype",
			input: "image/",
		},
		{
			name:  "missing slash",
			input: "image",
		},
		{
			name:  "missing type",
			input: "/jpeg",
		},
		{
			name:  "repeated slash",
			input: "image//jpeg",
		},
		{
			name:  "extra slash",
			input: "image/jpeg/extra",
		},
		{
			name:  "internal whitespace",
			input: "image /jpeg",
		},
		{
			name:  "whitespace around slash",
			input: "image/ jpeg",
		},
		{
			name:  "wildcard subtype",
			input: "image/*",
		},
		{
			name:  "wildcard type",
			input: "*/*",
		},
		{
			name:  "parameter charset",
			input: "image/jpeg; charset=utf-8",
		},
		{
			name:  "parameter boundary",
			input: "image/jpeg; boundary=test",
		},
		{
			name:  "malformed parameter",
			input: "image/jpeg; charset",
		},
		{
			name:  "multiple parameters",
			input: "image/jpeg; charset=utf-8; boundary=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMIME(tt.input)

			if err == nil {
				t.Fatalf("expected error for input %q", tt.input)
			}

			if got != "" {
				t.Errorf("expected empty result, got %q", got)
			}
		})
	}
}

// ------------------------------------------------------------
// Abnormal characters
// ------------------------------------------------------------

func TestNormalizeMIMEAbnormalCharacters(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "newline",
			input: "image/jpeg\n",
		},
		{
			name:  "carriage return",
			input: "image/jpeg\r",
		},
		{
			name:  "tab inside MIME",
			input: "image/\tjpeg",
		},
		{
			name:  "null byte",
			input: "image/jpeg\x00",
		},
		{
			name:  "control byte",
			input: "image/\x01jpeg",
		},
		{
			name:  "unicode subtype",
			input: "image/jpег",
		},
		{
			name:  "unicode type",
			input: "imаge/jpeg",
		},
		{
			name:  "invalid UTF-8",
			input: "image/\xff",
		},
		{
			name:  "extra semicolon",
			input: "image/jpeg;;",
		},
		{
			name:  "extra colon",
			input: "image/jpeg:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMIME(tt.input)

			if err == nil {
				t.Fatalf("expected error for input %q", tt.input)
			}

			if got != "" {
				t.Errorf("expected empty result, got %q", got)
			}
		})
	}
}

// ------------------------------------------------------------
// Length boundary
// ------------------------------------------------------------

func TestNormalizeMIMEBoundary(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "one byte",
			input: "i",
		},
		{
			name:  "maximum configured length",
			input: strings.Repeat("a", int(MaxMIMEbytes)),
		},
		{
			name:  "maximum length plus one",
			input: strings.Repeat("a", int(MaxMIMEbytes)+1),
		},
		{
			name:  "very long image prefix",
			input: "image/" + strings.Repeat("a", int(MaxMIMEbytes)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMIME(tt.input)

			if err == nil {
				t.Fatalf("expected error for input length %d", len(tt.input))
			}

			if got != "" {
				t.Errorf("expected empty result, got %q", got)
			}
		})
	}
}

// ------------------------------------------------------------
// Idempotence
// ------------------------------------------------------------

// Normalizing an already normalized MIME must not change its value.

func TestNormalizeMIMEIdempotence(t *testing.T) {
	inputs := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			first, err := NormalizeMIME(input)
			if err != nil {
				t.Fatalf("first normalization failed: %v", err)
			}

			second, err := NormalizeMIME(first)
			if err != nil {
				t.Fatalf("second normalization failed: %v", err)
			}

			if first != second {
				t.Fatalf(
					"normalization is not idempotent: first=%q second=%q",
					first,
					second,
				)
			}
		})
	}
}

// ------------------------------------------------------------
// Error output consistency
// ------------------------------------------------------------

func TestNormalizeMIMEErrorReturnsEmpty(t *testing.T) {
	inputs := []string{
		"",
		"application/pdf",
		"text/plain",
		"image/gif",
		"image/",
		"image//jpeg",
		"image/jpeg; charset=utf-8",
	}

	for _, input := range inputs {
		t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
			got, err := NormalizeMIME(input)

			if err == nil {
				t.Fatalf("expected error for %q", input)
			}

			if got != "" {
				t.Fatalf(
					"error must not return partial result: got %q",
					got,
				)
			}
		})
	}
}

// ------------------------------------------------------------
// Concurrency
// ------------------------------------------------------------

func TestNormalizeMIMEConcurrent(t *testing.T) {
	const (
		workers    = 100
		iterations = 100
	)

	inputs := []struct {
		input string
		want  string
	}{
		{"image/jpeg", "image/jpeg"},
		{"IMAGE/PNG", "image/png"},
		{"Image/WebP", "image/webp"},
	}

	var wg sync.WaitGroup

	errCh := make(chan error, workers*iterations)

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for i := 0; i < iterations; i++ {
				tt := inputs[i%len(inputs)]

				got, err := NormalizeMIME(tt.input)
				if err != nil {
					errCh <- fmt.Errorf(
						"worker %d: unexpected error: %w",
						workerID,
						err,
					)
					return
				}

				if got != tt.want {
					errCh <- fmt.Errorf(
						"worker %d: got %q, want %q",
						workerID,
						got,
						tt.want,
					)
					return
				}
			}
		}(worker)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

// ============================================================
// LEVEL 3 — EXTREME INPUT TESTS
// Input tidak normal sama sekali: fuzz dan input acak.
// ============================================================

func FuzzNormalizeMIME(f *testing.F) {
	seeds := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"IMAGE/JPEG",
		"Image/PnG",
		"application/pdf",
		"text/plain",
		"image/gif",
		"image/svg+xml",
		"",
		" ",
		"image",
		"image/",
		"image//jpeg",
		"image/jpeg/extra",
		"image/*",
		"*/*",
		"image/jpeg; charset=utf-8",
		"image/jpeg; charset",
		"image/jpег",
		"image/jpeg\x00",
		"image/\xff",
		"image/jpeg\n",
		strings.Repeat("a", int(MaxMIMEbytes)+1),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got, err := NormalizeMIME(input)

		// Error must never return a partial result.
		if err != nil {
			if got != "" {
				t.Fatalf(
					"error returned with non-empty result: %q",
					got,
				)
			}

			return
		}

		// Successful output must belong to the exact allowlist.
		switch got {
		case "image/jpeg", "image/png", "image/webp":
		default:
			t.Fatalf("unexpected normalized MIME: %q", got)
		}

		// Successful output must be idempotent.
		second, err := NormalizeMIME(got)
		if err != nil {
			t.Fatalf("normalized MIME rejected on second pass: %v", err)
		}

		if got != second {
			t.Fatalf(
				"normalization is not idempotent: %q != %q",
				got,
				second,
			)
		}

		// Successful output must not exceed configured length.
		if len(got) > int(MaxMIMEbytes) {
			t.Fatalf(
				"output exceeds maximum length: %d",
				len(got),
			)
		}
	})
}

// ============================================================
// BENCHMARK TESTS
// ============================================================

func BenchmarkNormalizeMIME(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "jpeg",
			input: "image/jpeg",
		},
		{
			name:  "png",
			input: "image/png",
		},
		{
			name:  "webp",
			input: "image/webp",
		},
		{
			name:  "uppercase",
			input: "IMAGE/JPEG",
		},
		{
			name:  "mixed_case",
			input: "Image/PnG",
		},
		{
			name:  "outer_whitespace",
			input: "  image/webp  ",
		},
		{
			name:  "unsupported_application",
			input: "application/pdf",
		},
		{
			name:  "unsupported_text",
			input: "text/plain",
		},
		{
			name:  "unsupported_image",
			input: "image/gif",
		},
		{
			name:  "malformed",
			input: "image//jpeg",
		},
		{
			name:  "parameter",
			input: "image/jpeg; charset=utf-8",
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = NormalizeMIME(tt.input)
			}
		})
	}
}
