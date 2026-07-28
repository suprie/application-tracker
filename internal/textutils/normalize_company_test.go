package textutils

import "testing"

func TestNormalizeCompanyName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple", "Apple", "apple"},
		{"lowercases", "MICROSOFT", "microsoft"},
		{"strips inc", "Apple Inc.", "apple"},
		{"strips llc", "Google LLC", "google"},
		{"strips corp punctuation and trims", "  Microsoft Corp.  ", "microsoft"},
		{"strips ltd", "Acme Ltd", "acme"},
		{"strips gmbh", "Siemens GmbH", "siemens"},
		{"strips se", "SAP SE", "sap"},
		{"strips multiple trailing suffixes", "Foo Bar LLC Inc", "foo bar"},
		{"preserves internal tokens", "Tata Consultancy Services Pvt Ltd", "tata consultancy services"},
		{"preserves internal co", "Co-op Bank", "co op bank"},
		{"keeps non-suffix trailing word", "Foo Ltd USA", "foo ltd usa"},
		{"collapses punctuation", "A.B.C. Holdings", "a b c holdings"},
		{"empty", "", ""},
		{"only suffix", "LLC", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeCompanyName(tt.in); got != tt.want {
				t.Errorf("NormalizeCompanyName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
