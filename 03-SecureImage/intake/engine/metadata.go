package engine

import (
	norm "03-SecureImage/intake/internal/normalize"
	in "03-SecureImage/intake"
)


func NormalizeImageMetadata(mime string, fileName string) ([2]in.ImageFormat, in.Status, error) {
	canonicalMIME, err := norm.NormalizeMIME(mime)
	if err != nil {
		return [2]in.ImageFormat{}, in.REJECT, err
	}

	normalizedFileName, err := norm.NormalizeFileName(fileName)
	if err != nil {
		return [2]in.ImageFormat{}, in.REJECT, err
	}

	return norm.ValidateMetadataConsistency(canonicalMIME, normalizedFileName)
}
