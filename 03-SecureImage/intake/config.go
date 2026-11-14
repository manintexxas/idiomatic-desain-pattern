package intake

type ImageConfig struct {
	MaxFileBytes		int64
	MaxImagePixel		int64	
	batchSize			int
	allowedForamt	  	[]ImageFormat
}
