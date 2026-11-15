package normalize

func NormalizeImageMetadata(mime string, fileName string) (string, string, error) {
	canonicalMIME, err := NormalizeMIME(mime)
	if err != nil {
		return "", "", err
	}

	normalizedFileName, err := NormalizeFileName(fileName)
	if err != nil {
		return "", "", err
	}

	if err := validateMetadataConsistency(canonicalMIME, normalizedFileName); err != nil {
		return "", "", err
	}

	return normalizedFileName, canonicalMIME, nil
}
