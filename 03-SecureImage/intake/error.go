package intake

import "errors"

var (
	ErrEmptyString		  = errors.New("Input cannot be empty")
	ErrInvalidConfig      = errors.New("Configure is invalid")
	ErrEmptyPayload       = errors.New("Payload is empty")
	ErrUnsupportedFormat  = errors.New("Unsupported Format")
	ErrMIMEFormatMismatch = errors.New("MIME doesn't match")
	ErrFileToolarge       = errors.New("File too large")
	ErrInvalidBatchSize   = errors.New("Batch size is invalid")
)
