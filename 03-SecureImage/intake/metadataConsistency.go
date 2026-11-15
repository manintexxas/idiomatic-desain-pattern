package intake

func NormalizeImageMetadata(mime string, fileName string) (string, string, error) {
	canonicalMIME, err := normalizeMIME(mime)
	if err != nil {
		return "", "", err
	}

	normalizedFileName, err := normalizeFileName(fileName)
	if err != nil {
		return "", "", err
	}

	if err := validateMetadataConsistency(canonicalMIME, normalizedFileName); err != nil {
		return "", "", err
	}

	return normalizedFileName, canonicalMIME, nil
}
