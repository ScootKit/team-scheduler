package utils

import "testing"

func TestIsApprovedEmailDomain(t *testing.T) {
	cases := []struct {
		name    string
		allowed string
		email   string
		want    bool
	}{
		{"empty allowlist permits all", "", "anyone@gmail.com", true},
		{"employee domain allowed", "scootkit.com,helper.scootkit.co", "alice@scootkit.com", true},
		{"helper domain allowed", "scootkit.com,helper.scootkit.co", "bob@helper.scootkit.co", true},
		{"case insensitive", "scootkit.com", "Bob@ScootKit.com", true},
		{"whitespace in list tolerated", " scootkit.com , helper.scootkit.co ", "c@helper.scootkit.co", true},
		{"disallowed domain rejected", "scootkit.com", "eve@gmail.com", false},
		{"subdomain is not the same domain", "scootkit.com", "x@evil.scootkit.com", false},
		{"helper not matched by employee domain", "scootkit.com", "h@helper.scootkit.co", false},
		{"no at sign rejected", "scootkit.com", "notanemail", false},
		{"empty domain rejected", "scootkit.com", "trailing@", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ALLOWED_EMAIL_DOMAINS", tc.allowed)
			if got := IsApprovedEmailDomain(tc.email); got != tc.want {
				t.Errorf("IsApprovedEmailDomain(%q) with allowlist %q = %v, want %v", tc.email, tc.allowed, got, tc.want)
			}
		})
	}
}
