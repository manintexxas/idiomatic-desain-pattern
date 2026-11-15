package engine

import (
	norm "03-SecureImage/intake/internal/normalize"
	in "03-SecureImage/intake"
)


func NormalizeImageMetadata(mime string, fileName string) in.Status {
	canonicalMIME, err := norm.NormalizeMIME(mime)
	if err != nil {
		return in.REJECT
	}

	normalizedFileName, err := norm.NormalizeFileName(fileName)
	if err != nil {
		return in.REJECT
	}

	return norm.ValidateMetadataConsistency(canonicalMIME, normalizedFileName)
}
