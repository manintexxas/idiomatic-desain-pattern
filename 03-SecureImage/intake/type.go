package intake

type (
	unknown string
	jpeg    string
	png     string
	webp    string
)

type imageFormat struct {
	formatUnknown unknown
	formatJPEG	  jpeg
	formatPNG	  png
	formatWebP	  webp
}

type imageUpload struct {
	fileName    string
	declareMIME string
	content     []byte
}

type acceptedImage struct {
	originalIndex	int
	fileName		string
	format			imageFormat
	content			[]byte
}

type rejectedImage struct {
	originalIndex	int
	fileName		string
	reason			error
}

type batchReport struct {
	accepted		[]acceptedImage
	rejected		[]rejectedImage
	bartches		[][]acceptedImage
}

type batchSummary struct {
	totalInput		int
	acceptedCount	int
	rejectedCount	int
	acceptedBytes	int64
	batchCount		int
}