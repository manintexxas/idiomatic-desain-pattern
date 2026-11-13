package intake

type (
	ImageFormat string
	Flag		uint8
)

var charTable[256]Flag

const (
	FormatUnknown	ImageFormat = "unknown"
	FormatJPEG		ImageFormat = "jpeg"
	FormatPNG		ImageFormat = "png"
	FormatWebP		ImageFormat = "webp"
)

const (
	Uppercase		Flag = 1 << iota
	Lowercase
	Symbolchar
	Whitespace
	Number
)

const (
	MaxFileNameBytes = 0xFF
	MaxASCIIChar     = 0x7F
	CharDot          = 0x2E
	CharHyphen       = 0x2D
	CharSlash        = 0x2F
	CharBackslash    = 0x5C
	CharUnderscore   = 0x5F
	CharMultipleDot  = 0x3A
)
type ImageUpload struct {
	FileName    string
	DeclareMIME string
	Content     []byte
}

type AcceptedImage struct {
	OriginalIndex	int
	FileName		string
	Format			ImageFormat
	Content			[]byte
}

type RejectedImage struct {
	OriginalIndex	int
	FileName		string
	Reason			error
}

type BatchReport struct {
	Accepted		[]AcceptedImage
	Rejected		[]RejectedImage
	Batches			[][]AcceptedImage
}

type BatchSummary struct {
	TotalInput		int
	AcceptedCount	int
	RejectedCount	int
	AcceptedBytes	int64
	BatchCount		int
}