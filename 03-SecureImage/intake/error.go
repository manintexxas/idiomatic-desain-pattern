
package intake

import "errors"

var (
    ErrEmptyMetadata					= errors.New("Metadata cannot be empty")

    ErrForbiddenCharacter				= errors.New("Contains forbidden character")
    ErrInvalidFileName					= errors.New("Invalid file name")
    ErrFileNameContainsInvalidSymbol 	= errors.New("Filename contains invalid symbol")
    ErrFileNameTooLong 					= errors.New("Filename exceeds maximum length")

    ErrInvalidMIME 						= errors.New("Invalid MIME")
    ErrMIMETooLong 						= errors.New("MIME exceeds maximum length")
    ErrMIMEContainsInvalidToken 		= errors.New("MIME contains invalid token")
    ErrMIMEParameterNotAllowed 			= errors.New("MIME parameter is not allowed")

    ErrMIMEMismatchFileName 			= errors.New("MIME and filename extension do not match")
    ErrMIMEMismatchSignature 			= errors.New("MIME and signature do not match")

    ErrInvalidConfig 					= errors.New("Configuration is invalid")

    ErrEmptyPayload 					= errors.New("Payload is empty")
    ErrUnsupportedFormat 				= errors.New("Unsupported format")
    ErrFileTooLarge 					= errors.New("File exceeds maximum size")
    ErrHeaderTooShort 					= errors.New("Signature is too short")
    ErrImagePixelLimitExceeded 			= errors.New("Image pixel limit exceeded")
    ErrDecodeFailed 					= errors.New("Image decoding failed")

    ErrBatchLimitExceeded 				= errors.New("Batch exceeds maximum file count")
)
