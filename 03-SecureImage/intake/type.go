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
	whitespace
	Number
)

const (
	MaxASCIIChar = 126
	CharDot      = '.'
	CharHyphen   = '-'
	CharSlash    = '/'
	CharBackslash = '\\'
	CharUnderscore = 95
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