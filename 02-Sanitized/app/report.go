package app

import (
	"fmt"
)

func (v *ViolationReport) HasViolations() bool {
	return v.TotalViolations > 0
}

func (v *ViolationReport) GetErrors(withCount bool) []string {
	rules := [...]struct {
		err   error
		count int
	}{
		{
			ErrNumberNotAllowed,
			v.NumberViolations},

		{
			ErrSymbolNotAllowed,
			v.SymbolViolations},

		{
			ErrWhitespaceOnly,
			v.WhitespaceViolations},

		{
			ErrUpperNotAllowed,
			v.UpperViolations},

		{
			ErrLowerNotAllowed,
			v.LowerViolations},
	}

	errs := make([]string, 0, len(rules))

	for _, r := range rules {
		if r.count > 0 {
			errs = append(errs, formatError(r.err, r.count, withCount))
		}
	}

	return errs
}

func formatError(err error, count int, withCount bool) string {
	if withCount {
		return fmt.Sprintf("%v: contains %d", err, count)
	}
	return err.Error()
}
