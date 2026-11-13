package intake

import (
	"fmt"
	"strings"
)

func NormalizeFileName(name string) ([]byte, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, fmt.Errorf("%w", ErrEmptyString)
	}

	if name[0] == '.' {
        return nil, fmt.Errorf("%w: file name cannot start with a dot", ErrInvalidPath)
    }

	b := []byte(name)

	for i := range b {
		char := b[i]

		if char > MaxASCIIChar {
			return nil, fmt.Errorf("%w: %s (%c) is multibyte", ErrMultibyteDetec, name, char)
		}

		if charTable[char] == 0 {
			return nil, fmt.Errorf("%w: %s contains forbidden char",ErrInvalidPath, name)
		}

		if char == CharSlash || char == CharBackslash {
			return nil, fmt.Errorf("%w: %c not allowed",ErrInvalidPath, char )
		}

		if char == CharDot && i > 0 && b[i-1] == CharDot {
            return nil, fmt.Errorf("%w: consecutive dots not allowed", ErrInvalidPath)
        }

		if charTable[char] == Uppercase {
			char |= 32
		}

		if shouldConvertToUnderscore(char) {
			char = CharUnderscore
		}

		b[i] = char
	}

	return b, nil

}

