package normalize

import (
		"strings"
		in "03-SecureImage/intake"
)

func trims(s string) string {
	start := 0
	for start < len(s) && s[start] == Space {
		start++
	}
	end := len(s)
	for end > start && s[end-1] == Space {
		end--
	}

	if start == 0 && end == len(s) {
		return s
	}

	if start == end {
		return ""
	}

	return strings.Clone(s[start:end])
}

func lower(s string) string {
	hasUpper := false
	for i := 0; i < len(s); i++ {
		if IsUpper(s[i]) {
			hasUpper = true
			break
		}
	}

	if !hasUpper {
		return s
	}

	buf := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if IsUpper(c) {
			c |= 32
		}

		buf[i] = c
	}

	return string(buf)
}

func length(name string) error {
	if name == "" {
		return in.ErrEmptyMetadata
	}

	return nil
}

func maxLen(name string, t Types) error {
	length := len(name)

	if t == File {
		if length > MaxFileNameBytes {
			return in.ErrFileNameTooLong
		}
	}

	if t == Mime {
		if length > MaxMIMEbytes {
			return in.ErrMIMETooLong
		}
	}

	return nil
}
