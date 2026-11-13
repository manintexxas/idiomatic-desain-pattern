package intake

import (
	"strings"
)

func NormalizeFileName(name string) (string, error) {
	name = cleanedInput(name)

	if err := validateFileName(name); err != nil {
		return "", err
	}

	return name, nil
}

func validateFileName(name string) error {
	if err := lenCheck(name, MaxFileNameBytes); err != nil {
		return err
	}

	for i := 0; i < len(name); i++ {
		char := name[i]

		switch {
		case char < 0x20 || char == 0x7F:
			return ErrInvalidFileName

		case char > MaxASCIIChar:
			return ErrInvalidFileName

		case char == CharBackslash || char == CharSlash || char == CharColon:
			return ErrForbbiden

		case !isAllowedFileNameByte(char):
			return ErrInvalidFileName

		case char == CharDot &&
			(i == 0 || i == len(name)-1 || name[i-1] == CharDot):
			return ErrForbbiden
		}
	}

	return nil
}

func NormalizeMIME(mime string) (string, error) {
	mime = cleanedInput(mime)

	if !strings.HasPrefix(mime, "image/") {
		return "", ErrMIMEFormatMismatch
	}

	if err := validateMIME(mime); err != nil {
		return "", err
	}

	subType := strings.TrimPrefix(mime, "image/")

	var format ImageFormat

	switch subType {
	case "jpeg", "jpg":
		format = FormatJPEG
		mime = "image/jpeg" 

	case "png":
		format = FormatPNG
		mime = "image/png"

	case "webp":
		format = FormatWebP
		mime = "image/webp"

	default:
		return "", ErrMIMEFormatMismatch
	}

	_ = format

	return mime, nil
}

func validateMIME(mime string) error {
	if err := lenCheck(mime, MaxMIMEbytes); err != nil {
		return err
	}

	slashCount := 0

	for i := 0; i < len(mime); i++ {
		char := mime[i]

		if charTable[char] == 0 || charTable[char] == Whitespace {
			return ErrForbbiden
		}

		if char > MaxASCIIChar {
			return ErrForbbiden
		}

		if char == CharSlash {
			slashCount++

			if slashCount > 1 {
				return ErrForbbiden
			}

			if i == 0 || i == len(mime)-1 {
				return ErrForbbiden
			}
		}

		if (charTable[char] & Symbolchar != 0) && !tokenCheck(char) {
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