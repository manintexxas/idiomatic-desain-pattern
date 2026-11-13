package intake

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// Three levels are ordered by attacker capability and potential impact.

// -----------------------------------------------------------------------------
// Shared assertions
// -----------------------------------------------------------------------------

func assertFileNameRejected(t *testing.T, input string) {
	t.Helper()

	got, err := NormalizeFileName(input)
	if err == nil {
		t.Fatalf("malicious filename accepted: input=%q output=%q", input, got)
	}
	if got != "" {
		t.Fatalf("partial output after error: input=%q output=%q err=%v", input, got, err)
	}
}

func assertMIMERejected(t *testing.T, mime, fileName string) {
	t.Helper()

	gotMIME, gotFormat, err := NormalizeMIME(mime, fileName)
	if err == nil {
		t.Fatalf(
			"malicious MIME accepted: mime=%q filename=%q output=(%q,%v)",
			mime, fileName, gotMIME, gotFormat,
		)
	}
	if gotMIME != "" {
		t.Errorf("MIME output after error = %q, want empty", gotMIME)
	}
	if gotFormat != FormatUnknown {
		t.Errorf("format after error = %v, want FormatUnknown", gotFormat)
	}
}

func assertErrorIs(t *testing.T, got, want error) {
	t.Helper()

	if !errors.Is(got, want) {
		t.Fatalf("error=%v, want errors.Is(_, %v)", got, want)
	}
}

// =============================================================================
// LEVEL 1 — DIRECT EXPLOITATION
// =============================================================================

func TestLevel1_NormalizeFileNameAcceptsOnlySupportedNames(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"jpeg", "photo.jpg", "photo.jpg"},
		{"jpeg long extension", "photo.jpeg", "photo.jpeg"},
		{"png", "photo.png", "photo.png"},
		{"webp", "photo.webp", "photo.webp"},
		{"uppercase", "PHOTO.JPG", "photo.jpg"},
		{"mixed case", "My_Photo-001.PnG", "my_photo-001.png"},
		{"outer whitespace", "  photo.jpg  ", "photo.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeFileName(tt.input)
			if err != nil {
				t.Fatalf("NormalizeFileName() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestLevel1_NormalizeMIMEAcceptsOnlySupportedPairs(t *testing.T) {
	tests := []struct {
		name       string
		mime       string
		fileName   string
		wantMIME   string
		wantFormat ImageFormat
	}{
		{"jpeg", "image/jpeg", "photo.jpg", "image/jpeg", FormatJPEG},
		{"jpg alias", "image/jpg", "photo.jpg", "image/jpeg", FormatJPEG},
		{"jpeg extension", "image/jpeg", "photo.jpeg", "image/jpeg", FormatJPEG},
		{"png", "image/png", "photo.png", "image/png", FormatPNG},
		{"webp", "image/webp", "photo.webp", "image/webp", FormatWebP},
		{"case normalization", "IMAGE/JPEG", "PHOTO.JPG", "image/jpeg", FormatJPEG},
		{"outer whitespace", " image/PNG ", " photo.png ", "image/png", FormatPNG},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMIME, gotFormat, err := NormalizeMIME(tt.mime, tt.fileName)
			if err != nil {
				t.Fatalf("NormalizeMIME() error = %v", err)
			}
			if gotMIME != tt.wantMIME || gotFormat != tt.wantFormat {
				t.Fatalf(
					"got=(%q,%v) want=(%q,%v)",
					gotMIME, gotFormat, tt.wantMIME, tt.wantFormat,
				)
			}
		})
	}
}

func TestLevel1_RejectsDirectPathAndFilenameAbuse(t *testing.T) {
	inputs := []struct {
		name  string
		value string
	}{
		{"unix traversal", "../photo.jpg"},
		{"nested unix traversal", "../../private/photo.jpg"},
		{"unix absolute path", "/tmp/photo.jpg"},
		{"windows traversal", `..\photo.jpg`},
		{"windows absolute path", `C:\temp\photo.jpg`},
		{"windows drive relative", `C:photo.jpg`},
		{"alternate data stream", "photo.jpg:payload"},
		{"double extension", "photo.jpg.exe"},
		{"unsupported extension", "photo.gif"},
		{"multiple dots", "photo..jpg"},
		{"empty basename", ".jpg"},
		{"empty extension", "photo."},
		{"internal whitespace", "my photo.jpg"},
		{"NUL byte", "photo\x00.jpg"},
		{"CRLF injection", "photo.jpg\r\nX-Injected: yes"},
	}

	for _, tt := range inputs {
		t.Run(tt.name, func(t *testing.T) {
			assertFileNameRejected(t, tt.value)
		})
	}
}

func TestLevel1_RejectsMIMEAndExtensionMismatch(t *testing.T) {
	pairs := []struct {
		name     string
		mime     string
		fileName string
	}{
		{"jpeg claims png", "image/jpeg", "photo.png"},
		{"png claims webp", "image/png", "photo.webp"},
		{"webp claims jpeg", "image/webp", "photo.jpg"},
		{"unsupported gif", "image/gif", "photo.gif"},
		{"application pdf", "application/pdf", "photo.jpg"},
		{"empty filename", "image/jpeg", ""},
		{"no extension", "image/jpeg", "photo"},
	}

	for _, tt := range pairs {
		t.Run(tt.name, func(t *testing.T) {
			assertMIMERejected(t, tt.mime, tt.fileName)
		})
	}
}

func TestLevel1_ErrorContractNeverReturnsPartialValues(t *testing.T) {
	badNames := []string{
		"", "../x.jpg", "x.gif", "x.jpg.exe", "x\x00.jpg",
	}

	for _, input := range badNames {
		t.Run("filename/"+fmt.Sprintf("%q", input), func(t *testing.T) {
			assertFileNameRejected(t, input)
		})
	}

	badPairs := [][2]string{
		{"image/jpeg", "../x.jpg"},
		{"image/jpeg", "x.png"},
		{"image/gif", "x.gif"},
		{"image/jpeg;foo=bar", "x.jpg"},
	}

	for _, pair := range badPairs {
		t.Run("mime/"+pair[0]+"/"+pair[1], func(t *testing.T) {
			assertMIMERejected(t, pair[0], pair[1])
		})
	}
}

// =============================================================================
// LEVEL 2 — EVASION AND INTERPRETATION DIFFERENTIALS
// =============================================================================

func TestLevel2_NormalizeMIMECannotBypassFilenameValidation(t *testing.T) {
	filenames := []struct {
		name  string
		value string
	}{
		{"unix traversal", "../photo.jpg"},
		{"windows traversal", `..\photo.jpg`},
		{"absolute windows path", `C:\photo.jpg`},
		{"drive relative path", `C:photo.jpg`},
		{"windows ADS", "photo.jpg:secret"},
		{"double extension", "photo.jpg.exe"},
		{"reserved CON", "CON.jpg"},
		{"reserved NUL", "NUL.png"},
		{"reserved AUX", "AUX.webp"},
		{"reserved COM1", "COM1.jpeg"},
		{"reserved LPT9", "LPT9.jpg"},
		{"CRLF suffix", "photo.jpg\r\nHeader: x"},
		{"NUL suffix", "photo.jpg\x00"},
	}

	for _, tt := range filenames {
		t.Run(tt.name, func(t *testing.T) {
			assertMIMERejected(t, "image/jpeg", tt.value)
		})
	}
}

func TestLevel2_RejectsUnicodeNormalizationAndVisualSpoofing(t *testing.T) {
	inputs := []struct {
		name  string
		value string
	}{
		{"Kelvin sign", "photo\u212A.jpg"},
		{"long s", "photo\u017F.jpg"},
		{"Cyrillic homoglyph", "photo.jpеg"},
		{"fullwidth dot", "photo\uFF0Ejpg"},
		{"right-to-left override", "photo\u202Egpj.jpg"},
		{"zero width space", "photo\u200B.jpg"},
		{"combining mark", "photo\u0301.jpg"},
		{"non-breaking space", "photo\u00A0.jpg"},
		{"emoji", "photo\U0001F4A5.jpg"},
		{
			"malformed UTF-8",
			string([]byte{'p', 'h', 'o', 't', 'o', 0xff, '.', 'j', 'p', 'g'}),
		},
	}

	for _, tt := range inputs {
		t.Run(tt.name, func(t *testing.T) {
			assertFileNameRejected(t, tt.value)
		})
	}
}

func TestLevel2_RejectsMIMESyntaxEvasion(t *testing.T) {
	mimes := []struct {
		name  string
		value string
	}{
		{"missing subtype", "image/"},
		{"missing slash", "image"},
		{"multiple slash", "image//jpeg"},
		{"trailing slash", "image/jpeg/"},
		{"wildcard subtype", "image/*"},
		{"wildcard type", "*/jpeg"},
		{"charset parameter", "image/jpeg;charset=utf-8"},
		{"boundary parameter", "image/jpeg;boundary=test"},
		{"parameter whitespace", "image/jpeg; charset=utf-8"},
		{"quoted parameter", `image/jpeg;name="photo"`},
		{"CRLF injection", "image/jpeg\r\nX-Injected: yes"},
		{"embedded NUL", "image/jpe\x00g"},
		{"tab injection", "image/\tjpeg"},
		{"Unicode subtype", "image/jpеg"},
		{
			"invalid UTF-8",
			string([]byte{'i', 'm', 'a', 'g', 'e', '/', 'j', 'p', 0xff, 'g'}),
		},
	}

	for _, tt := range mimes {
		t.Run(tt.name, func(t *testing.T) {
			assertMIMERejected(t, tt.value, "photo.jpg")
		})
	}
}

func TestLevel2_ReservedWindowsDeviceNames(t *testing.T) {
	reserved := []string{
		"CON", "con", "PrN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5",
		"COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5",
		"LPT6", "LPT7", "LPT8", "LPT9",
		"CLOCK$", "CONIN$", "CONOUT$",
	}

	for _, base := range reserved {
		t.Run(base, func(t *testing.T) {
			if !isReservedWindowsDeviceName(base) {
				t.Fatalf("reserved name not detected: %q", base)
			}
		})
	}

	ordinary := []string{
		"photo", "CONSOLE", "COM0", "COM10",
		"LPT0", "LPT10", "AUXILIARY", "NULFILE",
	}

	for _, base := range ordinary {
		t.Run("not_reserved_"+base, func(t *testing.T) {
			if isReservedWindowsDeviceName(base) {
				t.Fatalf("ordinary name incorrectly reserved: %q", base)
			}
		})
	}
}

func TestLevel2_DotAndSuffixAmbiguity(t *testing.T) {
	inputs := []string{
		".photo.jpg",
		"photo..jpg",
		"photo...jpg",
		"photo.jpg.",
		"photo.jpg..",
		"photo.jpg.exe",
		"photo.png.jpg",
		"photo.jpg%00",
		"photo%2ejpg",
		"photo.jpg::$DATA",
		"photo.jpg:stream",
		"photo.jpg/..",
		`photo.jpg\..`,
	}

	for _, input := range inputs {
		t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
			assertFileNameRejected(t, input)
		})
	}
}

func TestLevel2_WhitespaceNormalizationHasStableOutput(t *testing.T) {
	valid := []struct {
		input string
		want  string
	}{
		{" photo.jpg", "photo.jpg"},
		{"photo.jpg ", "photo.jpg"},
		{"\tphoto.jpg\t", "photo.jpg"},
	}

	for _, tt := range valid {
		got, err := NormalizeFileName(tt.input)
		if err != nil {
			t.Fatalf("NormalizeFileName(%q): %v", tt.input, err)
		}
		if got != tt.want {
			t.Fatalf("got=%q want=%q", got, tt.want)
		}
	}

	assertFileNameRejected(t, "photo .jpg")
	assertFileNameRejected(t, "photo\t.jpg")
}
// =============================================================================
// LEVEL 3 — FATAL-INPUT RESILIENCE
// =============================================================================

func TestLevel3_AllByteValuesDoNotPanicAndUnsafeBytesAreRejected(t *testing.T) {
	for value := 0; value <= 0xff; value++ {
		b := byte(value)

		fileName := string([]byte{'a', b, '.', 'j', 'p', 'g'})
		gotName, nameErr := NormalizeFileName(fileName)

		if value < 0x20 || value == 0x7f || value > int(MaxASCIIChar) {
			if nameErr == nil {
				t.Fatalf("filename accepted forbidden byte 0x%02x: %q", value, gotName)
			}
		}

		if nameErr != nil && gotName != "" {
			t.Fatalf("partial filename output for byte 0x%02x: %q", value, gotName)
		}

		mime := string(append([]byte("image/jpe"), b, 'g'))
		gotMIME, gotFormat, mimeErr := NormalizeMIME(mime, "photo.jpg")

		if value < 0x20 || value == 0x7f || value > int(MaxASCIIChar) {
			if mimeErr == nil {
				t.Fatalf("MIME accepted forbidden byte 0x%02x", value)
			}
		}

		if mimeErr != nil && (gotMIME != "" || gotFormat != FormatUnknown) {
			t.Fatalf("partial MIME output for byte 0x%02x", value)
		}
	}
}

func TestLevel3_RejectsEveryHighBitByteInFilename(t *testing.T) {
	for value := 0x80; value <= 0xff; value++ {
		input := string([]byte{'a', byte(value), '.', 'j', 'p', 'g'})

		got, err := NormalizeFileName(input)
		if err == nil {
			t.Fatalf("high-bit byte 0x%02x accepted: %q", value, got)
		}
		if got != "" {
			t.Fatalf("partial output for high-bit byte 0x%02x: %q", value, got)
		}
	}
}

func TestLevel3_RawLengthBoundaries(t *testing.T) {
	maxName := int(MaxFileNameBytes)
	if maxName < len("a.jpg") {
		t.Skipf("MaxFileNameBytes=%d too small", maxName)
	}

	atNameLimit := strings.Repeat("a", maxName-len(".jpg")) + ".jpg"

	got, err := NormalizeFileName(atNameLimit)
	if err != nil {
		t.Fatalf("filename at limit rejected: %v", err)
	}
	if len(got) != maxName {
		t.Fatalf("filename bytes=%d want=%d", len(got), maxName)
	}

	aboveNameLimit := atNameLimit + "a"
	assertFileNameRejected(t, aboveNameLimit)

	maxMIME := int(MaxMIMEbytes)
	if maxMIME < len("image/jpeg") {
		t.Skipf("MaxMIMEbytes=%d too small", maxMIME)
	}

	atMIMELimit := "a/" + strings.Repeat("b", maxMIME-2)
	if err := validateMIME(atMIMELimit); err != nil {
		t.Fatalf("MIME at limit rejected: %v", err)
	}

	aboveMIMELimit := "a/" + strings.Repeat("b", maxMIME-1)
	if err := validateMIME(aboveMIMELimit); err == nil {
		t.Fatal("MIME above byte limit accepted")
	}

	assertMIMERejected(
		t,
		strings.Repeat("x", maxMIME+1),
		"photo.jpg",
	)
}

func TestLevel3_MegabyteInputsAreRejectedByConfiguredCaps(t *testing.T) {
	nameSize := 1 << 20
	if nameSize <= int(MaxFileNameBytes) {
		nameSize = int(MaxFileNameBytes) + 1
	}

	mimeSize := 1 << 20
	if mimeSize <= int(MaxMIMEbytes) {
		mimeSize = int(MaxMIMEbytes) + 1
	}

	assertFileNameRejected(t, strings.Repeat(" ", nameSize))
	assertMIMERejected(t, strings.Repeat("x", mimeSize), "photo.jpg")
}

func TestLevel3_OversizedWhitespaceIsRejectedBeforeCleanup(t *testing.T) {
	name := strings.Repeat(" ", int(MaxFileNameBytes)+1)
	assertFileNameRejected(t, name)

	mime := strings.Repeat(" ", int(MaxMIMEbytes)+1)
	assertMIMERejected(t, mime, "photo.jpg")
}

func TestLevel3_ConcurrentMixedTrustedAndHostileInputs(t *testing.T) {
	const workers = 32
	const iterations = 200

	type inputCase struct {
		name     string
		mime     string
		fileName string
		valid    bool
	}

	cases := []inputCase{
		{"jpeg", "image/jpeg", "photo.jpg", true},
		{"png", "image/png", "photo.png", true},
		{"webp", "image/webp", "photo.webp", true},
		{"path traversal", "image/jpeg", "../photo.jpg", false},
		{"windows ADS", "image/jpeg", "photo.jpg:stream", false},
		{"reserved device", "image/jpeg", "CON.jpg", false},
		{"MIME mismatch", "image/png", "photo.jpg", false},
		{"parameter injection", "image/jpeg;foo=bar", "photo.jpg", false},
		{"unicode filename", "image/jpeg", "ph\u212Foto.jpg", false},
	}

	start := make(chan struct{})
	errCh := make(chan error, workers)

	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			<-start

			for iteration := 0; iteration < iterations; iteration++ {
				for _, tc := range cases {
					gotMIME, gotFormat, err := NormalizeMIME(tc.mime, tc.fileName)

					if tc.valid {
						if err != nil {
							errCh <- fmt.Errorf(
								"worker=%d case=%s unexpected error: %w",
								workerID, tc.name, err,
							)
							return
						}

						if gotMIME == "" || gotFormat == FormatUnknown {
							errCh <- fmt.Errorf(
								"worker=%d case=%s incomplete result",
								workerID, tc.name,
							)
							return
						}
					} else {
						if err == nil ||
							gotMIME != "" ||
							gotFormat != FormatUnknown {
							errCh <- fmt.Errorf(
								"worker=%d case=%s hostile input accepted",
								workerID, tc.name,
							)
							return
						}
					}
				}
			}
		}(worker)
	}

	close(start)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

func TestLevel3_RepeatedMalformedInputsRemainStable(t *testing.T) {
	inputs := []struct {
		mime     string
		fileName string
	}{
		{"image/jpeg\r\nX: y", "photo.jpg"},
		{"image/jpeg", "../photo.jpg"},
		{"image/jpe\x00g", "photo.jpg"},
		{"image/png", "photo.jpg"},
		{"image/*", "photo.jpg"},
	}

	for i := 0; i < 5000; i++ {
		for _, tc := range inputs {
			gotMIME, gotFormat, err := NormalizeMIME(tc.mime, tc.fileName)

			if err == nil {
				t.Fatalf(
					"iteration=%d malicious input accepted: %q %q",
					i, tc.mime, tc.fileName,
				)
			}

			if gotMIME != "" || gotFormat != FormatUnknown {
				t.Fatalf(
					"iteration=%d partial output: (%q,%v)",
					i, gotMIME, gotFormat,
				)
			}
		}
	}
}

// =============================================================================
// DIRECT HELPER TESTS — FUNCTION COVERAGE
// =============================================================================

func TestValidateFileNameBranches(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid jpg", "photo.jpg", nil},
		{"valid jpeg", "photo.jpeg", nil},
		{"valid png", "photo.png", nil},
		{"valid webp", "photo.webp", nil},
		{"upper extension", "PHOTO.JPG", nil},
		{"empty", "", ErrEmptyString},
		{"no dot", "photo", ErrInvalidFileName},
		{"unsupported extension", "photo.gif", ErrInvalidFileName},
		{"multiple dots", "photo..jpg", ErrInvalidFileName},
		{"leading dot", ".jpg", ErrForbbiden},
		{"slash", "folder/photo.jpg", ErrForbbiden},
		{"backslash", `folder\photo.jpg`, ErrForbbiden},
		{"colon", "C:photo.jpg", ErrForbbiden},
		{"control", "photo\n.jpg", ErrInvalidFileName},
		{"delete", "photo\x7f.jpg", ErrInvalidFileName},
		{
			"high bit",
			string([]byte{'p', 0xff, '.', 'j', 'p', 'g'}),
			ErrInvalidFileName,
		},
		{"invalid punctuation", "photo@.jpg", ErrInvalidFileName},
		{"reserved CON", "CON.jpg", ErrForbbiden},
		{"reserved COM1", "COM1.jpg", ErrForbbiden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFileName(tt.input)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("validateFileName(%q): %v", tt.input, err)
				}
				return
			}

			assertErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestValidateFileNameRejectsOversizedRawValue(t *testing.T) {
	input := strings.Repeat("a", int(MaxFileNameBytes)+1)

	if err := validateFileName(input); err == nil {
		t.Fatal("validateFileName accepted oversized input")
	}
}

func TestValidateMIMEBranches(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"jpeg", "image/jpeg", nil},
		{"png", "image/png", nil},
		{"webp", "image/webp", nil},
		{"generic application", "application/pdf", nil},
		{"generic text", "text/plain", nil},
		{"hyphen token", "application/x-test", nil},
		{"plus token", "application/x-test+json", nil},
		{"empty", "", ErrEmptyString},
		{"missing slash", "imagejpeg", ErrForbbiden},
		{"slash first", "/jpeg", ErrForbbiden},
		{"slash last", "image/", ErrForbbiden},
		{"multiple slash", "image//jpeg", ErrForbbiden},
		{"whitespace", "image/j peg", ErrForbbiden},
		{"control", "image/jpe\x00g", ErrForbbiden},
		{
			"high bit",
			string([]byte{'i', 'm', 'a', 'g', 'e', '/', 0xff}),
			ErrForbbiden,
		},
		{"disallowed punctuation", "image/jpeg;foo=bar", ErrForbbiden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMIME(tt.input)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("validateMIME(%q): %v", tt.input, err)
				}
				return
			}

			assertErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestIsAllowedImageExtension(t *testing.T) {
	allowed := []string{
		".jpg", ".jpeg", ".png", ".webp", ".JPG", ".PnG",
	}

	for _, extension := range allowed {
		if !isAllowedImageExtension(extension) {
			t.Errorf("extension %q should be allowed", extension)
		}
	}

	rejected := []string{
		"", ".gif", ".svg", ".jpg.exe", "jpg", ".jpeg/", ".jpg\\",
	}

	for _, extension := range rejected {
		if isAllowedImageExtension(extension) {
			t.Errorf("extension %q should be rejected", extension)
		}
	}
}

func TestMIMEMatchesExtension(t *testing.T) {
	tests := []struct {
		format    ImageFormat
		extension string
		want      bool
	}{
		{FormatJPEG, ".jpg", true},
		{FormatJPEG, ".jpeg", true},
		{FormatJPEG, ".JPG", true},
		{FormatJPEG, ".png", false},
		{FormatPNG, ".png", true},
		{FormatPNG, ".PNG", true},
		{FormatPNG, ".jpg", false},
		{FormatWebP, ".webp", true},
		{FormatWebP, ".jpg", false},
		{FormatUnknown, ".jpg", false},
	}

	for _, tt := range tests {
		got := mimeMatchesExtension(tt.format, tt.extension)
		if got != tt.want {
			t.Errorf(
				"mimeMatchesExtension(%v,%q)=%v want=%v",
				tt.format, tt.extension, got, tt.want,
			)
		}
	}
}

func TestContainsNonASCIIByte(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"photo.jpg", false},
		{"image/jpeg", false},
		{"", false},
		{"photo-é.jpg", true},
		{"photo\u212A.jpg", true},
		{string([]byte{0xff}), true},
	}

	for _, tt := range tests {
		if got := containsNonASCIIByte(tt.input); got != tt.want {
			t.Errorf(
				"containsNonASCIIByte(%q)=%v want=%v",
				tt.input, got, tt.want,
			)
		}
	}
}

// =============================================================================
// PROPERTY / REGRESSION TESTS
// =============================================================================

func TestNormalizeFileNameIsIdempotent(t *testing.T) {
	inputs := []string{
		"photo.jpg",
		"PHOTO.JPG",
		"  My_Photo-001.PNG  ",
		"image.webp",
	}

	for _, input := range inputs {
		t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
			first, err := NormalizeFileName(input)
			if err != nil {
				t.Fatalf("first normalization failed: %v", err)
			}

			second, err := NormalizeFileName(first)
			if err != nil {
				t.Fatalf("second normalization failed: %v", err)
			}

			if second != first {
				t.Fatalf("not idempotent: first=%q second=%q", first, second)
			}
		})
	}
}

func TestNormalizeMIMEIsIdempotent(t *testing.T) {
	inputs := []struct {
		mime     string
		fileName string
	}{
		{"image/jpeg", "photo.jpg"},
		{"image/jpg", "photo.jpeg"},
		{"image/png", "photo.png"},
		{"image/webp", "photo.webp"},
	}

	for _, input := range inputs {
		t.Run(input.mime+"/"+input.fileName, func(t *testing.T) {
			firstMIME, firstFormat, err := NormalizeMIME(input.mime, input.fileName)
			if err != nil {
				t.Fatalf("first normalization failed: %v", err)
			}

			secondMIME, secondFormat, err := NormalizeMIME(firstMIME, input.fileName)
			if err != nil {
				t.Fatalf("second normalization failed: %v", err)
			}

			if firstMIME != secondMIME || firstFormat != secondFormat {
				t.Fatalf(
					"not idempotent: first=(%q,%v) second=(%q,%v)",
					firstMIME, firstFormat, secondMIME, secondFormat,
				)
			}
		})
	}
}

func TestSuccessfulFilenameAlwaysSatisfiesValidator(t *testing.T) {
	inputs := []string{
		"photo.jpg",
		"PHOTO.PNG",
		"  image_01.webp  ",
		"test-file.jpeg",
	}

	for _, input := range inputs {
		got, err := NormalizeFileName(input)
		if err != nil {
			t.Fatalf("NormalizeFileName(%q): %v", input, err)
		}

		if err := validateFileName(got); err != nil {
			t.Fatalf("successful output %q fails validator: %v", got, err)
		}
	}
}

func TestNormalizeMIMERejectsDirectCallerFilenameBypass(t *testing.T) {
	for _, fileName := range []string{
		"../photo.jpg",
		`..\photo.jpg`,
		`C:\temp\photo.jpg`,
		"photo.jpg:stream",
		"CON.jpg",
		"photo.jpg.exe",
	} {
		assertMIMERejected(t, "image/jpeg", fileName)
	}
}

// =============================================================================
// FUZZ TESTS
//
// Versioned target names isolate new corpora from legacy fuzz targets
// that may have a different number of parameters.
// =============================================================================

// -----------------------------------------------------------------------------
// Fuzz filename normalization
// -----------------------------------------------------------------------------

func FuzzNormalizeFileName(f *testing.F) {
	seeds := []string{
		"photo.jpg",
		"PHOTO.PNG",
		"  photo.webp  ",
		"photo.jpeg",
		"",
		"../photo.jpg",
		`..\photo.jpg`,
		`C:\photo.jpg`,
		"photo.jpg.exe",
		"photo\x00.jpg",
		"photo\r\nX-Test: yes.jpg",
		"photo\u212A.jpg",
		"photo\u202Egpj.jpg",
		"CON.jpg",
		"COM1.jpg",
		string([]byte{0xff, 0xfe, 0x00, '.', 'j', 'p', 'g'}),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > int(MaxFileNameBytes)*2+1024 {
			t.Skip()
		}

		got, err := NormalizeFileName(input)

		if err != nil {
			if got != "" {
				t.Fatalf("error result must be empty: got=%q err=%v", got, err)
			}
			return
		}

		if len(got) > int(MaxFileNameBytes) {
			t.Fatalf(
				"successful output exceeds byte limit: %d > %d",
				len(got), MaxFileNameBytes,
			)
		}

		if err := validateFileName(got); err != nil {
			t.Fatalf("successful output fails validator: %q: %v", got, err)
		}

		if containsNonASCIIByte(got) {
			t.Fatalf("successful output contains non-ASCII byte: %q", got)
		}

		again, err := NormalizeFileName(got)
		if err != nil {
			t.Fatalf("normalizing successful output again failed: %v", err)
		}

		if again != got {
			t.Fatalf("not idempotent: first=%q second=%q", got, again)
		}
	})
}

// -----------------------------------------------------------------------------
// Fuzz MIME input — one argument
// -----------------------------------------------------------------------------

func FuzzNormalizeMIMEInputV2(f *testing.F) {
	seeds := []string{
		"image/jpeg",
		"image/jpg",
		"IMAGE/JPEG",
		"image/png",
		"image/webp",
		"application/pdf",
		"image/gif",
		"image/",
		"image//jpeg",
		"image/*",
		"image/jpeg;foo=bar",
		"image/jpeg\r\nX-Test: yes",
		string([]byte{'i', 'm', 'a', 'g', 'e', '/', 0xff}),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, mime string) {
		if len(mime) > int(MaxMIMEbytes)*2+1024 {
			t.Skip()
		}

		assertNormalizeMIMEResultInvariant(t, mime, "fuzz.jpg")
	})
}

// -----------------------------------------------------------------------------
// Fuzz MIME + filename pair
// -----------------------------------------------------------------------------

func FuzzNormalizeMIMEFilenamePair(f *testing.F) {
	seeds := [][2]string{
		{"image/jpeg", "photo.jpg"},
		{"image/jpg", "photo.jpeg"},
		{"image/png", "photo.png"},
		{"image/webp", "photo.webp"},
		{"IMAGE/JPEG", "PHOTO.JPG"},
		{"application/pdf", "photo.jpg"},
		{"image/jpeg", "../photo.jpg"},
		{"image/jpeg", `..\photo.jpg`},
		{"image/jpeg", "CON.jpg"},
		{"image/jpeg", "photo.jpg.exe"},
		{"image/png", "photo.jpg"},
		{"image/jpeg;foo=bar", "photo.jpg"},
		{"image/jpeg\r\nX: y", "photo.jpg"},
	}

	for _, seed := range seeds {
		f.Add(seed[0], seed[1])
	}

	f.Fuzz(func(t *testing.T, mime, fileName string) {
		if len(mime) > int(MaxMIMEbytes)*2+1024 ||
			len(fileName) > int(MaxFileNameBytes)*2+1024 {
			t.Skip()
		}

		assertNormalizeMIMEResultInvariant(t, mime, fileName)
	})
}

// -----------------------------------------------------------------------------
// Fuzz internal filename validator
// -----------------------------------------------------------------------------

func FuzzValidateFileName(f *testing.F) {
	seeds := []string{
		"photo.jpg",
		"PHOTO.PNG",
		"photo.webp",
		"",
		".jpg",
		"photo..jpg",
		"../photo.jpg",
		`..\photo.jpg`,
		"CON.jpg",
		"COM1.jpg",
		"photo.jpg:stream",
		string([]byte{'p', 0xff, '.', 'j', 'p', 'g'}),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > int(MaxFileNameBytes)*2+1024 {
			t.Skip()
		}

		if err := validateFileName(input); err != nil {
			return
		}

		if len(input) > int(MaxFileNameBytes) {
			t.Fatalf("validator accepted oversized filename: %d bytes", len(input))
		}

		if containsNonASCIIByte(input) {
			t.Fatalf("validator accepted non-ASCII filename: %q", input)
		}

		if strings.ContainsAny(input, `/\\:`) {
			t.Fatalf("validator accepted path/ADS delimiter: %q", input)
		}

		dotIndex := strings.LastIndexByte(input, CharDot)
		if dotIndex <= 0 {
			t.Fatalf("validator accepted invalid extension boundary: %q", input)
		}

		if isReservedWindowsDeviceName(input[:dotIndex]) {
			t.Fatalf("validator accepted reserved device basename: %q", input)
		}
	})
}

// -----------------------------------------------------------------------------
// Fuzz internal MIME validator
// -----------------------------------------------------------------------------

func FuzzValidateMIME(f *testing.F) {
	seeds := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"application/pdf",
		"text/plain",
		"",
		"image/",
		"/jpeg",
		"image//jpeg",
		"image/j peg",
		"image/jpeg; charset=utf-8",
		"image/*",
		string([]byte{'i', 'm', 'a', 'g', 'e', '/', 0xff}),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > int(MaxMIMEbytes)*2+1024 {
			t.Skip()
		}

		if err := validateMIME(input); err != nil {
			return
		}

		if len(input) > int(MaxMIMEbytes) {
			t.Fatalf("validator accepted oversized MIME: %d bytes", len(input))
		}

		slashCount := 0

		for i := 0; i < len(input); i++ {
			char := input[i]

			if char > MaxASCIIChar || char < 0x20 || char == 0x7f {
				t.Fatalf("validator accepted forbidden byte 0x%02x", char)
			}

			if isWhiteSpace(char) {
				t.Fatalf("validator accepted whitespace byte 0x%02x", char)
			}

			if char == CharSlash {
				slashCount++
			}
		}

		if slashCount != 1 {
			t.Fatalf("validator accepted %d slashes, want exactly one", slashCount)
		}
	})
}

// -----------------------------------------------------------------------------
// MIME output invariants
// -----------------------------------------------------------------------------

func assertNormalizeMIMEResultInvariant(
	t *testing.T,
	mime string,
	fileName string,
) {
	t.Helper()

	gotMIME, gotFormat, err := NormalizeMIME(mime, fileName)

	if err != nil {
		if gotMIME != "" || gotFormat != FormatUnknown {
			t.Fatalf(
				"error result must be empty/unknown: (%q,%v), err=%v",
				gotMIME, gotFormat, err,
			)
		}
		return
	}

	if err := validateMIME(gotMIME); err != nil {
		t.Fatalf("successful MIME output rejected by validator: %q: %v", gotMIME, err)
	}

	if gotFormat == FormatUnknown {
		t.Fatal("successful normalization returned FormatUnknown")
	}

	var expected ImageFormat

	switch gotMIME {
	case "image/jpeg":
		expected = FormatJPEG
	case "image/png":
		expected = FormatPNG
	case "image/webp":
		expected = FormatWebP
	default:
		t.Fatalf("unexpected canonical MIME: %q", gotMIME)
	}

	if gotFormat != expected {
		t.Fatalf(
			"MIME/format invariant violated: MIME=%q format=%v want=%v",
			gotMIME, gotFormat, expected,
		)
	}

	normalizedName, nameErr := NormalizeFileName(fileName)
	if nameErr != nil {
		t.Fatalf(
			"MIME normalization succeeded for rejected filename %q: %v",
			fileName, nameErr,
		)
	}

	dotIndex := strings.LastIndexByte(normalizedName, CharDot)
	if dotIndex <= 0 || dotIndex >= len(normalizedName)-1 {
		t.Fatalf("successful filename has invalid extension boundary: %q", normalizedName)
	}

	if !mimeMatchesExtension(gotFormat, normalizedName[dotIndex:]) {
		t.Fatalf(
			"successful MIME/extension mismatch: MIME=%q filename=%q",
			gotMIME, normalizedName,
		)
	}
}

// =============================================================================
// BENCHMARKS
//
// Jalankan dengan -benchmem untuk mendapatkan:
// ns/op    : waktu eksekusi
// B/op     : byte allocation
// allocs/op: jumlah allocation per operasi
// =============================================================================

var (
	benchmarkNormalizeString string
	benchmarkNormalizeFormat ImageFormat
	benchmarkNormalizeError  error
)

// -----------------------------------------------------------------------------
// NormalizeFileName
// -----------------------------------------------------------------------------

func BenchmarkNormalizeFileName(b *testing.B) {
	maxBytes := int(MaxFileNameBytes)
	maxName := "photo.jpg"

	if maxBytes >= len("a.jpg") {
		maxName = strings.Repeat("a", maxBytes-len(".jpg")) + ".jpg"
	}

	cases := []struct {
		name  string
		input string
	}{
		{"short_valid", "photo.jpg"},
		{"case_and_trim", "  PHOTO_001.JPG  "},
		{"max_length", maxName},
		{"invalid_extension", "photo.gif"},
		{"path_traversal", "../photo.jpg"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tc.input)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				benchmarkNormalizeString, benchmarkNormalizeError =
					NormalizeFileName(tc.input)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// validateFileName
// -----------------------------------------------------------------------------

func BenchmarkValidateFileName(b *testing.B) {
	cases := []struct {
		name  string
		input string
	}{
		{"valid", "photo_001.jpg"},
		{"invalid_extension", "photo.gif"},
		{"traversal", "../photo.jpg"},
		{"reserved_device", "CON.jpg"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tc.input)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				benchmarkNormalizeError = validateFileName(tc.input)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// NormalizeMIME
// -----------------------------------------------------------------------------

func BenchmarkNormalizeMIME(b *testing.B) {
	cases := []struct {
		name     string
		mime     string
		fileName string
	}{
		{"jpeg", "image/jpeg", "photo.jpg"},
		{"png", "image/png", "photo.png"},
		{"webp", "image/webp", "photo.webp"},
		{"jpg_alias", "image/jpg", "photo.jpeg"},
		{"mismatch", "image/jpeg", "photo.png"},
		{"direct_traversal", "image/jpeg", "../photo.jpg"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tc.mime) + len(tc.fileName)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				benchmarkNormalizeString,
					benchmarkNormalizeFormat,
					benchmarkNormalizeError =
					NormalizeMIME(tc.mime, tc.fileName)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// validateMIME
// -----------------------------------------------------------------------------

func BenchmarkValidateMIME(b *testing.B) {
	cases := []struct {
		name  string
		input string
	}{
		{"valid", "image/jpeg"},
		{"invalid_parameter", "image/jpeg;foo=bar"},
		{"invalid_slash", "image//jpeg"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(tc.input)))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				benchmarkNormalizeError = validateMIME(tc.input)
			}
		})
	}
}