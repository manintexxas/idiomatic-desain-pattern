package intake

type (
	ImageFormat string
	Status uint8
)

const (
	FormatUnknown	ImageFormat = ""
	FormatJPEG		ImageFormat = "jpeg"
	FormatJPG		ImageFormat = "jpg"
	FormatPNG		ImageFormat = "png"
	FormatWebP		ImageFormat = "webp"
)

const (
	PASS 	Status = 1 << iota
	FLAG
	REJECT
)

var ext = [4]string{"jpg", "jpeg", "png", "webp"}

type ImageUpload struct {
	FileName    	string
	DeclaredMIME 	string
	Content     	[]byte
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