package app

import (
	"fmt"
)

type ViolationItem struct {
	Cause	  string `json:"Cause"`
	Contains int	 `json:"Contains"`
}

type DetailBreakdown struct {
	TotalViolations int `json:"Total"`
	Violations		[]ViolationItem `json:"Violations"`
}

func (v *ViolationReport) HasViolations() bool {
	return v.TotalViolations > 0
}

func (v *ViolationReport) TriggerMalformedEarlyExit() {
	*v = ViolationReport{
			MalformedUnicodeViolations: 1,
			TotalViolations: 			1,
	}
}

func (v *ViolationReport) Reset() {
	*v = ViolationReport{}
}

func (v *ViolationReport) GetErrorsDetail() DetailBreakdown {
	rules := [...]struct {
		err   error
		count int
	}{
		{ErrMalformedUnicode, 
		v.MalformedUnicodeViolations},

		{ErrDisallowedKeyword,
		v.KeywordViolations},

		{ErrLengthLess, 
		v.MinLengthViolations},
		
		{ErrLengthMore, 
		v.MaxLengthViolations},

		{ErrMinLetterViolation, 
		v.MinLetterViolations},

		{ErrMaxLetterViolation, 
		v.MaxLetterViolations},

		{ErrMinUpperViolation, 
		v.MinUpperViolations},

		{ErrMaxUpperViolation, 
		v.MaxUpperViolations},

		{ErrMinLowerViolation, 
		v.MinLowerViolations},

		{ErrMaxLowerViolation, 
		v.MaxLowerViolations},

		{ErrMinNumberViolation, 
		v.MinNumberViolations},

		{ErrMaxNumberViolation, 
		v.MaxNumberViolations},

		{ErrMinSymbolViolation, 
		v.MinSymbolViolations},

		{ErrMaxSymbolViolation, 
		v.MaxSymbolViolations},

		{ErrMinWhitespaceViolation, 
		v.MinWhiteSpaceViolations},

		{ErrMaxWhitespaceViolation, 
		v.MaxWhiteSpaceViolations},

		{ErrUpperNotAllowed, 
		v.UpperViolations},
		{
			ErrLowerNotAllowed, 
		v.LowerViolations},
		{
			ErrNumberNotAllowed, 
		v.NumberViolations},
		
		{ErrSymbolNotAllowed, 
		v.SymbolViolations},
		
		{ErrWhitespaceOnly, 
		v.WhitespaceViolations},
	}

	items := make([]ViolationItem, 0, len(rules))

	for _, r := range rules {
		if r.count > 0 {
			items = append(items, ViolationItem{
					Cause: r.err.Error(),
					Contains: r.count,
			})
		}
	}

	return DetailBreakdown{
		TotalViolations: 	v.TotalViolations,
		Violations: 		items,
	}
}

