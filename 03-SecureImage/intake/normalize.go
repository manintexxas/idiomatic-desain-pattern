
package intake

import (
	"fmt"
	"strings"
)

func NormalizeFileName(name string) (string, error) {
	name = strings.Trim(name, " ")

	if err := validateFileName(name); err != nil {
		return "", err
	}

	return strings.ToLower(name), nil
}

func ValidateFileName(name string) error {
	name = strings.Trim(name, " ")

	return validateFileName(name)
}

func validateFileName(name string) error {
	if name == "" {
		return ErrEmptyString
	}

	if len(name) > MaxFileNameBytes {
		return fmt.Errorf(
			"%w: maximum %d bytes, got %d",
			ErrFileNameTooLong,
			MaxFileNameBytes,
			len(name),
		)
	}

	for i := 0; i < len(name); i++ {
		char := name[i]

		switch {
		case char < 0x20 || char == 0x7F:
			return fmt.Errorf(
				"%w: control character 0x%02X",
				ErrInvalidFileName,
				char,
			)

		case char > MaxASCIIChar:
			return fmt.Errorf(
				"%w: non-ASCII byte 0x%02X",
				ErrInvalidFileName,
				char,
			)

		case char == CharBackslash || char == CharSlash || char == CharMultipleDot:
			return fmt.Errorf(
				"%w: path character %q",
				ErrInvalidPath,
				char,
			)

		case !isAllowedFileNameByte(char):
			return fmt.Errorf(
				"%w: forbidden character %q",
				ErrInvalidFileName,
				char,
			)

		case char == CharDot &&
			(i == 0 || i == len(name)-1 || name[i-1] == CharDot):
			return fmt.Errorf(
				"%w: invalid dot placement",
				ErrInvalidPath,
			)
		}
	}

	return nil
}

