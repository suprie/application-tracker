package textutils

import (
	"regexp"
	"strings"
)

var companyTokenRe = regexp.MustCompile(`[a-z0-9]+`)

// companySuffixes are trailing legal-entity designations stripped during
// normalization (Inc, Ltd, LLC, Corp, GmbH, ...). They are matched only when
// they appear as the final token(s) so internal words like "Co-op" are kept.
var companySuffixes = map[string]bool{
	"inc": true, "llc": true, "ltd": true, "limited": true,
	"corp": true, "corporation": true, "co": true, "gmbh": true,
	"ag": true, "se": true, "sa": true, "sas": true, "sarl": true,
	"bv": true, "nv": true, "pvt": true, "plc": true, "llp": true,
	"lp": true, "spa": true, "srl": true, "oy": true, "ab": true,
	"kg": true, "pc": true, "pa": true,
}

// NormalizeCompanyName lowercases a company name, strips trailing legal
// suffixes, drops punctuation, and collapses whitespace. It is used for
// deduplication and lookup, not for display.
func NormalizeCompanyName(raw string) string {
	tokens := companyTokenRe.FindAllString(strings.ToLower(strings.TrimSpace(raw)), -1)

	// Strip trailing legal suffixes (handles e.g. "Foo Bar LLC Inc").
	for len(tokens) > 0 && companySuffixes[tokens[len(tokens)-1]] {
		tokens = tokens[:len(tokens)-1]
	}

	return strings.Join(tokens, " ")
}
