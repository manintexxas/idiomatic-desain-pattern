package normalize

import ("strings"
		in "03-SecureImage/intake"
)

func NormalizeMIME(mime string) (string, error) {
	if err := maxLen(mime, Mime); err != nil {
		return "", err
	}

	if err := baseCharCheck(mime); err != nil {
		return "", err
	}

	mime = lower(mime)

	if err := validateMIME(mime); err != nil {
		return "", err
	}

	if !strings.HasPrefix(mime, "image/") {
		return "", in.ErrInvalidMIME
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
		return "", in.ErrInvalidMIME
	}

	return canonicalMIME, nil
}

func validateMIME(mime string) error {
	slashCount := 0

	for i := 0; i < len(mime); i++ {
		char := mime[i]

		if char > MaxASCIIChar ||
			in.CharTable[char] == 0 ||
			isWhiteSpace(char) {
			return in.ErrForbiddenCharacter
		}

		if char == CharSlash {
			slashCount++

			if slashCount > 1 || i == 0 || i == len(mime)-1 {
				return in.ErrForbiddenCharacter
			}

			continue
		}

		if IsSymbol(char) && !IsAllowedMIMESymbol(char) {
			return in.ErrMIMEContainsInvalidToken
		}
	}

	if slashCount != 1 {
		return in.ErrForbiddenCharacter
	}

	return nil
}
