package intake

import (
	"path/filepath"
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

	if err := lenCheck(name, int(MaxFileNameBytes)); err != nil {
		return err
	}

	dotCount := 0
	for i := 0; i < len(name); i++ {
		if name[i] == CharDot {
			dotCount++
		}
	}

	if dotCount != 1 {
		return ErrInvalidFileName
	}

	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return ErrInvalidFileName
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

		case char == CharDot && (i == 0 || i == len(name)-1):
			return ErrForbbiden
		}
	}

	return nil
}

func NormalizeMIME(mime string, fileName string) (string, ImageFormat, error) {
	mime = cleanedInput(mime)

	if !strings.HasPrefix(mime, "image/") {
		return "", FormatUnknown, ErrMIMEFormatMismatch
	}

	if err := validateMIME(mime); err != nil {
		return "", FormatUnknown, err
	}

	subType := strings.TrimPrefix(mime, "image/")
	var format ImageFormat
	var canonicalMime string

	switch subType {
	case "jpeg", "jpg":
		format = FormatJPEG
		canonicalMime = "image/jpeg"

	case "png":
		format = FormatPNG
		canonicalMime = "image/png"

	case "webp":
		format = FormatWebP
		canonicalMime = "image/webp"

	default:
		return "", FormatUnknown, ErrMIMEFormatMismatch
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	match := false
	
	switch format {
	case FormatJPEG:
		match = (ext == ".jpg" || ext == ".jpeg")

	case FormatPNG:
		match = (ext == ".png")

	case FormatWebP:
		match = (ext == ".webp")
	}

	if !match {
		return "", FormatUnknown, ErrMIMEFormatMismatch
	}

	return canonicalMime, format, nil
}

func validateMIME(mime string) error {
	if err := lenCheck(mime, int(MaxMIMEbytes)); err != nil {
		return err
	}

	slashCount := 0

	for i := 0; i < len(mime); i++ {
		char := mime[i]

		if charTable[char] == 0 || isWhiteSpace(char) {
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