
package intake

import "strings"

func NormalizeFileName(name string) (string, error) {
	name = trimInput(name)

	if err := baseCharCheck(name, File); err != nil {
		return "", err
	}

	name = toLower(name)

	name = strings.ReplaceAll( name, string(Space), string(CharUnderscore))

	if err := validateFileName(name); err != nil {
		return "", err
	}

	return name, nil
}

func NormalizeMIME(mime string) (string, error) {
	if err := baseCharCheck(mime, Mime); err != nil {
		return "", err
	}

	mime = toLower(mime)

	if err := validateMIME(mime); err != nil {
		return "", err
	}

	if !strings.HasPrefix(mime, "image/") {
		return "", ErrMIMEFormatMismatch
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
		return "", ErrMIMEFormatMismatch
	}

	return canonicalMIME, nil
}

func NormalizeImageMetadata( mime string, fileName string ) (string, string, error) {
	canonicalMIME, err := NormalizeMIME(mime)
	if err != nil {
		return "", "", err
	}

	normalizedFileName, err := NormalizeFileName(fileName)
	if err != nil {
		return "", "", err
	}

	if err := validateMetadataConsistency( canonicalMIME, normalizedFileName ); err != nil {
		return "", "", err
	}

	return normalizedFileName, canonicalMIME, nil
}

func validateFileName(name string) error {
	dotIndex := -1
	dotCount := 0

	for i := 0; i < len(name); i++ {
		char := name[i]

		switch {

		case isWhiteSpace(char):
			return ErrForbbiden

		case !isAllowedFileNameByte(char):
			return ErrForbbiden
		}

		if char == CharDot {
			dotCount++
			dotIndex = i
		}
	}

	if dotCount != 1 {
		return ErrInvalidFileName
	}

	if dotIndex == 0 {
		return ErrForbbiden
	}

	if dotIndex == len(name)-1 {
		return ErrInvalidFileName
	}

	baseName := name[:dotIndex]
	extension := name[dotIndex:]

	if !isAllowedImageExtension(extension) {
		return ErrInvalidFileName
	}

	if isReservedWindowsDeviceName(baseName) {
		return ErrForbbiden
	}

	return nil
}

func validateMIME(mime string) error {
	slashCount := 0

	for i := 0; i < len(mime); i++ {
		char := mime[i]

		if char > MaxASCIIChar ||
			charTable[char] == 0 ||
			isWhiteSpace(char) {
			return ErrForbbiden
		}

		if char == CharSlash {
			slashCount++

			if slashCount > 1 || i == 0 || i == len(mime)-1 {
				return ErrForbbiden
			}

			continue
		}

		if IsSymbol(char) && !isAllowedMIMESymbol(char) {
			return ErrForbbiden
		}
	}

	if slashCount != 1 {
		return ErrForbbiden
	}

	return nil
}
