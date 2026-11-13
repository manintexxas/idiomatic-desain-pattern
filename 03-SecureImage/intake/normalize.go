package intake

import "strings"

func NormalizeFileName(name string) (string, error) {
	if err := lenCheck(name, int(MaxFileNameBytes)); err != nil {
		return "", err
	}

	if containsNonASCIIByte(name) {
		return "", ErrInvalidFileName
	}

	name = cleanedInput(name)

	if err := validateFileName(name); err != nil {
		return "", err
	}

	return name, nil
}

func validateFileName(name string) error {
	if err := lenCheck(name, int(MaxFileNameBytes)); err != nil {
		return err
	}

	if name == "" {
		return ErrInvalidFileName
	}

	dotIndex := -1
	dotCount := 0

	for i := 0; i < len(name); i++ {
		char := name[i]

		switch {
		case char > MaxASCIIChar:
			return ErrInvalidFileName

		case char < 0x20 || char == 0x7F:
			return ErrInvalidFileName

		case char == CharBackslash ||
			char == CharSlash ||
			char == CharColon:
			return ErrForbbiden

		case !isAllowedFileNameByte(char):
			return ErrInvalidFileName
		}

		if char == CharDot {
			dotCount++
			dotIndex = i
		}
	}

	if dotCount != 1 {
		return ErrInvalidFileName
	}

	baseName := name[:dotIndex]
	extension := name[dotIndex:]

	if !isAllowedImageExtension(extension) {
		return ErrInvalidFileName
	}

	if dotIndex == 0 {
		return ErrForbbiden
	}

	if isReservedWindowsDeviceName(baseName) {
		return ErrForbbiden
	}

	return nil
}

func NormalizeMIME(mime string, fileName string) (string, ImageFormat, error) {

	if err := lenCheck(mime, int(MaxMIMEbytes)); err != nil {
		return "", FormatUnknown, err
	}

	if containsNonASCIIByte(mime) {
		return "", FormatUnknown, ErrForbbiden
	}

	mime = cleanedInput(mime)

	if !strings.HasPrefix(mime, "image/") {
		return "", FormatUnknown, ErrMIMEFormatMismatch
	}

	if err := validateMIME(mime); err != nil {
		return "", FormatUnknown, err
	}

	subType := strings.TrimPrefix(mime, "image/")

	var format ImageFormat
	var canonicalMIME string

	switch subType {
	case "jpeg", "jpg":
		format = FormatJPEG
		canonicalMIME = "image/jpeg"

	case "png":
		format = FormatPNG
		canonicalMIME = "image/png"

	case "webp":
		format = FormatWebP
		canonicalMIME = "image/webp"

	default:
		return "", FormatUnknown, ErrMIMEFormatMismatch
	}

	normalizedFileName, err := NormalizeFileName(fileName)
	if err != nil {
		return "", FormatUnknown, ErrMIMEFormatMismatch
	}

	dotIndex := strings.LastIndexByte(
		normalizedFileName,
		CharDot,
	)

	if dotIndex <= 0 || dotIndex == len(normalizedFileName)-1 {
		return "", FormatUnknown, ErrMIMEFormatMismatch
	}

	extension := normalizedFileName[dotIndex:]

	if !mimeMatchesExtension(format, extension) {
		return "", FormatUnknown, ErrMIMEFormatMismatch
	}

	return canonicalMIME, format, nil
}

func validateMIME(mime string) error {
	if err := lenCheck(mime, int(MaxMIMEbytes)); err != nil {
		return err
	}

	if mime == "" {
		return ErrForbbiden
	}

	slashCount := 0

	for i := 0; i < len(mime); i++ {
		char := mime[i]

		if char > MaxASCIIChar ||
			char < 0x20 ||
			char == 0x7F {
			return ErrForbbiden
		}

		if charTable[char] == 0 || isWhiteSpace(char) {
			return ErrForbbiden
		}

		if char == CharSlash {
			slashCount++

			if slashCount > 1 ||
				i == 0 ||
				i == len(mime)-1 {
				return ErrForbbiden
			}

			continue
		}

		if IsSymbol(char) && !tokenCheck(char) {
			if char != CharPlus && char != CharHyphen {
				return ErrForbbiden
			}
		}
	}

	if slashCount != 1 {
		return ErrForbbiden
	}

	return nil
}