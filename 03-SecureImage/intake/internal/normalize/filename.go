package normalize

import ("strings"
		"03-SecureImage/intake"

func FileNamePatternCheck(s string) error {
	n := len(s)

	if n < 5 {
		return ErrFileNameTooShort
	}

	if !IsAllowedFileNameByte(s[0]) {
		return ErrForbiddenCharacter
	}

	if s[0] == CharDot {
		if s[1] == CharDot {
			return ErrUnnecessaryDot
		}

		return ErrMalformedMetadata
	}

	limit := max(n-1-MaxExtScan, 1)

	firstDot := -1
	previousDot := -1
	lastDot := -1

	for i := 1; i < n; i++ {
		c := s[i]

		if !IsAllowedFileNameByte(c) {
			return ErrForbiddenCharacter
		}

		if c != CharDot {
			continue
		}

		if s[i-1] == CharDot {
			return ErrUnnecessaryDot
		}

		if firstDot == -1 {
			firstDot = i

			if IsReservedWindowsDeviceName(s[:firstDot]) {
				return ErrMalformedMetadata
			}
		} else {
			start := previousDot + 1
			candidateLen := i - start

			if candidateLen <= MaxExtScan {
				candidate := s[start:i]

				if IsAllowedExt(candidate) == nil ||
					isDangerousExt(candidate) {
					return ErrMalformedMetadata
				}
			}
		}

		previousDot = i

		if i >= limit {
			lastDot = i
		}
	}

	if firstDot <= 0 || lastDot <= 0 || lastDot >= n-1 {
		return ErrMalformedMetadata
	}

	extStart := lastDot + 1

	for i := extStart; i < n; i++ {
		if s[i] == '(' || s[i] = ')' {
			return ErrForbiddenCharacter
		}
	}

	ext := s[extStart:]

	if err := IsAllowedExt(ext); err != nil {
		return err
	}

	return nil
}

func normalizeFileName(name string) (string, error) {
	if err := maxBytes(name, File); err != nil {
		return "", err
	}

	name = trimInput(name)

	if err := baseCharCheck(name); err != nil {
		return "", err
	}

	name = toLower(name)

	name = strings.ReplaceAll(name, string(Space), string(CharUnderscore))

	if err := fileNamePatternCheck(name); err != nil {
		return "", err
	}

	return name, nil
}
