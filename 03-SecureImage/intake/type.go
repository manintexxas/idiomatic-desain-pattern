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
	MaxFileNameBytes byte	= 0xFF
	MaxMIMEbytes	 byte	= 0x80
	MaxASCIIChar     byte 	= 0x7F
	CharDot          byte	= 0x2E
	CharHyphen       byte	= 0x2D
	CharSlash        byte	= 0x2F
	CharBackslash    byte	= 0x5C
	CharUnderscore   byte	= 0x5F
	CharColon		 byte	= 0x3A
	CharSemiColon	 byte	= 0x3B
	CharPlus  		 byte 	= 0x2B
    CharHash  		 byte 	= 0x23
    CharMod   		 byte 	= 0x25
    CharAnd   		 byte 	= 0x26
    CharQuote 		 byte 	= 0x27
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