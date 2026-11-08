package app

type CostumeRules struct {
	MinLength 		int
	MaxLength 		int
	MaxUpper 		int
	MaxLower 		int
	MaxNumbers 		int
	MaxSymbols 		int
	MaxWhiteSpace 	int
	ReqUpper 		int
	ReqLower 		int
	ReqNumbers 		int
	ReqSymbols 		int
	ReqWhiteSpace	int
}

func DefaultRules() CostumeRules {
	return CostumeRules{
		MinLength: 		0,
		MaxLength: 		0,
		MaxUpper: 		-1,
		MaxLower: 		-1,
		MaxNumbers: 	-1,
		MaxSymbols: 	-1,
		MaxWhiteSpace: 	-1,
		ReqUpper: 		0,
		ReqLower: 		0,
		ReqNumbers: 	0,
		ReqSymbols: 	0,
	}
}

func (cs *CostumeRules) ValidateRulesConfig () (error, bool) {
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

func ValidateRules(b []byte, cs CostumeRules, c *Config, r *ViolationReport) {
	if _, isInvalid := cs.ValidateRulesConfig(); isInvalid{
		r.TotalViolations = 1
		return
	}

	length := len(b)

	if cs.MinLength > 0 && length < cs.MinLength {
		r.MinLengthViolations++
	}

	if cs.MaxLength > 0 && length > cs.MaxLength {
		r.MaxLengthViolations++
	}

	if length == 0 {
		return
	}

	upper, lower, num, sym, ws := 0, 0, 0, 0, 0

	for index := 0; index < length; index++ {
		flag := CharTable[b[index]]


		switch {
		case ( flag & FlagAllowUpperOnly ) != 0:
			upper++
		
		case ( flag & FlagAllowUpperOnly ) != 0:
			lower++

		case ( flag & FlagAllowNumbers ) != 0:
			num++

		case ( flag & FlagAllowSymbols) != 0:
			sym++

		case ( flag & FlagAllowWhitespace ) != 0:
			ws++
		}
	}

	if ( c.UniversalFlags & FlagAllowUpperOnly ) == 0 && upper > 0 {
		r.UpperViolations += upper
	} else {
		if cs.MaxUpper >= 0 && upper > cs.MaxUpper {
			r.UpperViolations += ( upper - cs.MaxUpper)
		}
		if cs.ReqUpper > 0 && upper < cs.ReqUpper {
			r.UpperViolations += ( upper - cs.ReqUpper)
		}
	}
	

	if ( c.UniversalFlags & FlagAllowLowerOnly ) == 0 && lower > 0 {
		r.UpperViolations += upper
	} else {
		if cs.MinLength >= 0 && lower > cs.MaxUpper {
			r.LowerViolations += ( lower - cs.MaxLower )
		}
		if cs.ReqLower >= 0 && lower < cs.ReqLower {
			r.LowerViolations += ( lower - cs.ReqLower)
		}
	}

	if ( c.UniversalFlags & FlagAllowNumbers ) == 0 && lower > 0 {
		r.NumberViolations += num
	} else {
		if cs.MaxNumbers >= 0 && num > cs.MaxNumbers {
			r.NumberViolations += ( num - cs.MaxNumbers )
		}
		if cs.ReqNumbers >= 0 && num < cs.ReqNumbers {
			r.NumberViolations += ( num - cs.ReqNumbers )
		}
	}

	if ( c.UniversalFlags & FlagAllowSymbols ) == 0 && sym > 0 {
		r.SymbolViolations += sym 
	} else {
		if cs.MaxSymbols >= 0 && sym > cs.MaxSymbols {
			r.SymbolViolations += ( sym - cs.MaxSymbols )
		} 
		if cs.ReqSymbols >= 0 && sym < cs.ReqSymbols {
			r.SymbolViolations += ( sym - cs.ReqSymbols )
		}
	}

	if ( c.UniversalFlags & FlagAllowWhitespace ) == 0 && ws > 0 {
		r.WhitespaceViolations += ws
	} else {
		if cs.MaxWhiteSpace >= 0 && ws > cs.MaxWhiteSpace {
			r.WhitespaceViolations += ( ws - cs.MaxWhiteSpace )
		}
		if cs.ReqWhiteSpace >= 0 && ws < cs.ReqWhiteSpace {
			r.WhitespaceViolations += ( ws - cs.ReqWhiteSpace )
		}
	}

	
}