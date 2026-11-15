package intake

import "strings"

func normalizeMIME(mime string) (string, error) {
	if err := maxBytes(mime, Mime); err != nil {
		return "", err
	}

	if err := baseCharCheck(mime); err != nil {
		return "", err
	}

	mime = toLower(mime)

	if err := validateMIME(mime); err != nil {
		return "", err
	}

	if !strings.HasPrefix(mime, "image/") {
		return "", ErrInvalidMIME
	}

	subType := strings.TrimPrefix(mime, "image/")

	var canonicalMIME string

	switch subType {
	case "jpeg", "jpg":
		canonicalMIME = "image/jpeg"

	case "png":
		canonicalMIME = "image/png"

	case "webp":
		canonicalMIME = "image/webp"

	default:
		return "", ErrInvalidMIME
	}

	return canonicalMIME, nil
}

func validateMIME(mime string) error {
	slashCount := 0

	for i := 0; i < len(mime); i++ {
		char := mime[i]

		if char > MaxASCIIChar ||
			charTable[char] == 0 ||
			isWhiteSpace(char) {
			return ErrForbiddenCharacter
		}

		if char == CharSlash {
			slashCount++

			if slashCount > 1 || i == 0 || i == len(mime)-1 {
				return ErrForbiddenCharacter
			}

			continue
		}

		if IsSymbol(char) && !IsAllowedMIMESymbol(char) {
			return ErrMIMEContainsInvalidToken
		}
	}

	if slashCount != 1 {
		return ErrForbiddenCharacter
	}

	return nil
}
