package app

type CostumeRules struct {
	MinLength     int
	MaxLength     int
	MaxLetter     int
	MaxUpper      int
	MaxLower      int
	MaxNumbers    int
	MaxSymbols    int
	MaxWhiteSpace int
	ReqLetter     int
	ReqUpper      int
	ReqLower      int
	ReqNumbers    int
	ReqSymbols    int
	ReqWhiteSpace int
}

func DefaultRules() CostumeRules {
	return CostumeRules{
		MinLength:     0,
		MaxLength:     0,
		MaxLetter:     -1,
		MaxUpper:      -1,
		MaxLower:      -1,
		MaxNumbers:    -1,
		MaxSymbols:    -1,
		MaxWhiteSpace: -1,
		ReqLetter:     0,
		ReqUpper:      0,
		ReqLower:      0,
		ReqNumbers:    0,
		ReqSymbols:    0,
		ReqWhiteSpace: 0,
	}
}

func (cs *CostumeRules) ValidateRulesConfig() (error, bool) {
	if cs.MaxLetter >= 0 && cs.ReqLetter > cs.MaxLetter {
		return ErrReqMoreMax, true
	}
	if cs.MaxUpper >= 0 && cs.ReqUpper > cs.MaxUpper {
		return ErrReqMoreMax, true
	}
	if cs.MaxLower >= 0 && cs.ReqLower > cs.MaxLower {
		return ErrReqMoreMax, true
	}
	if cs.MaxNumbers >= 0 && cs.ReqNumbers > cs.MaxNumbers {
		return ErrReqMoreMax, true
	}
	if cs.MaxSymbols >= 0 && cs.ReqSymbols > cs.MaxSymbols {
		return ErrReqMoreMax, true
	}
	if cs.MaxWhiteSpace >= 0 && cs.ReqWhiteSpace > cs.MaxWhiteSpace {
		return ErrReqMoreMax, true
	}
	return nil, false
}

func ValidateRules(b []byte, cs CostumeRules, c *Config, e *Engine, r *ViolationReport) {
	if _, isInvalid := cs.ValidateRulesConfig(); isInvalid {
		r.TotalViolations = 1
		return
	}

	length := len(b)

	if cs.MinLength > 0 && length < cs.MinLength {
		diff := cs.MinLength - length
		r.MinLengthViolations = diff
		r.TotalViolations += diff
	}

	if cs.MaxLength > 0 && length > cs.MaxLength {
		diff := length - cs.MaxLength
		r.MaxLengthViolations = diff
		r.TotalViolations += diff
	}

	if length == 0 {
		return
	}

	upper, lower, num, sym, ws := 0, 0, 0, 0, 0

	for i := 0; i < length; i++ {
		flag := charTable[b[i]]

		if flag == 0 {
			r.MalformedUnicodeViolations = 1
			r.TotalViolations = 1
			return
		}

		switch {
		case (flag & FlagAllowUpperOnly) != 0:
			upper++
		case (flag & FlagAllowLowerOnly) != 0:
			lower++
		case (flag & FlagAllowNumbers) != 0:
			num++
		case (flag & FlagAllowSymbols) != 0:
			sym++
		case (flag & FlagAllowWhitespace) != 0:
			ws++
		}
	}

	letter := upper + lower

	if (c.UniversalFlags & LetterPreset) == 0 {
		if letter > 0 {
			r.LetterViolations += letter
			r.TotalViolations += letter
		}
	} else {
		if (c.UniversalFlags & FlagAllowUpperOnly) == 0 && upper > 0 {
			r.UpperViolations += upper
			r.TotalViolations += upper
		} else {
			if cs.MaxUpper >= 0 && upper > cs.MaxUpper {
				diff := upper - cs.MaxUpper
				r.MaxUpperViolations += diff
				r.TotalViolations += diff
			}
			if cs.ReqUpper > 0 && upper < cs.ReqUpper {
				diff := cs.ReqUpper - upper
				r.MinUpperViolations += diff
				r.TotalViolations += diff
			}
		}

		if (c.UniversalFlags & FlagAllowLowerOnly) == 0 && lower > 0 {
			r.LowerViolations += lower
			r.TotalViolations += lower
		} else {
			if cs.MaxLower >= 0 && lower > cs.MaxLower {
				diff := cs.MaxLower - lower
				r.MaxLowerViolations += diff
				r.TotalViolations += diff
			}
			if cs.ReqLower > 0 && lower < cs.ReqLower {
				diff := cs.ReqLower - lower
				r.MinLowerViolations += diff
				r.TotalViolations += diff
			}
		}

		if cs.MaxLetter >= 0 && letter > cs.MaxLetter {
			diff := letter - cs.MaxLetter
			r.MaxLetterViolations += diff
			r.TotalViolations += diff
		}
		if cs.ReqLetter > 0 && letter < cs.ReqLetter {
			diff := cs.ReqLetter - letter
			r.MinLetterViolations += diff
			r.TotalViolations += diff
		}
	}

	if (c.UniversalFlags & FlagAllowNumbers) == 0 && num > 0 {
		r.NumberViolations += num
		r.TotalViolations += num
	} else {
		if cs.MaxNumbers >= 0 && num > cs.MaxNumbers {
			diff := num - cs.MaxNumbers
			r.MaxNumberViolations += diff
			r.TotalViolations += diff
		}
		if cs.ReqNumbers > 0 && num < cs.ReqNumbers {
			diff := cs.ReqNumbers - num
			r.MinNumberViolations += diff
			r.TotalViolations += diff
		}
	}

	if (c.UniversalFlags & FlagAllowSymbols) == 0 && sym > 0 {
		r.SymbolViolations += sym
		r.TotalViolations += sym
	} else {
		if cs.MaxSymbols >= 0 && sym > cs.MaxSymbols {
			diff := sym - cs.MaxSymbols
			r.MaxSymbolViolations += diff
			r.TotalViolations += diff
		}
		if cs.ReqSymbols > 0 && sym < cs.ReqSymbols {
			diff := cs.ReqSymbols - sym
			r.MinSymbolViolations += diff
			r.TotalViolations += diff
		}
	}

	if (c.UniversalFlags & FlagAllowWhitespace) == 0 && ws > 0 {
		r.WhitespaceViolations += ws
		r.TotalViolations += ws
	} else {
		if cs.MaxWhiteSpace >= 0 && ws > cs.MaxWhiteSpace {
			diff := ws - cs.MaxWhiteSpace
			r.MaxWhiteSpaceViolations += diff
			r.TotalViolations += diff
		}
		if cs.ReqWhiteSpace > 0 && ws < cs.ReqWhiteSpace {
			diff := cs.ReqWhiteSpace - ws
			r.MinWhiteSpaceViolations += diff
			r.TotalViolations += diff
		}
	}

	if (c.UniversalFlags & FlagAllowKeywords) != 0 && e != nil && e.Trie != nil {
		if e.Trie.hasMatch(b) {
			r.KeywordViolations = 1
			r.TotalViolations++
		}
	}
}