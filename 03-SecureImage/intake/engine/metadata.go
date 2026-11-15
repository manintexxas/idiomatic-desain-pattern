package engine

import norm "03-SecureImage/intake/internal/normalize"

func NormalizeImageMetadata(mime string, fileName string) (string, string, error) {
	canonicalMIME, err := norm.NormalizeMIME(mime)
	if err != nil {
		return "", "", err
	}

	normalizedFileName, err := norm.NormalizeFileName(fileName)
	if err != nil {
		return "", "", err
	}

	if err := norm.ValidateMetadataConsistency(canonicalMIME, normalizedFileName); err != nil {
		return "", "", err
	}

	return normalizedFileName, canonicalMIME, nil
}
