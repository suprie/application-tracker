package domain

import "strings"

// NormalizeStatus maps a status alias (case-insensitive) to its canonical
// constant. The empty string maps to "" (i.e. no filter). ok is false for
// unrecognized values. Shared by the CLI list filter and the API status query.
func NormalizeStatus(s string) (status string, ok bool) {
	switch strings.ToLower(s) {
	case "draft":
		return StatusDraft, true
	case "fit match", "fitmatch", "fit_match":
		return StatusFitMatch, true
	case "applied":
		return StatusApplied, true
	case "rejected":
		return StatusRejected, true
	case "offer":
		return StatusOffer, true
	case "":
		return "", true
	default:
		return "", false
	}
}
