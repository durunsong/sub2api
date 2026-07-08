//go:build unit

package accessban

import "testing"

func TestMatchesClient(t *testing.T) {
	t.Parallel()
	if !MatchesClient(RuleTypeIP, "203.0.113.10", "", "203.0.113.10", "curl/8.0") {
		t.Fatal("expected ip match")
	}
	if MatchesClient(RuleTypeIP, "203.0.113.10", "", "203.0.113.11", "curl/8.0") {
		t.Fatal("expected ip mismatch")
	}
	if !MatchesClient(RuleTypeUA, "curl", "", "1.1.1.1", "curl/8.0") {
		t.Fatal("expected ua match")
	}
	if MatchesClient(RuleTypeIPUA, "203.0.113.10", "curl", "203.0.113.10", "wget/1.0") {
		t.Fatal("expected ip+ua mismatch on ua")
	}
	if !MatchesClient(RuleTypeIPUA, "203.0.113.10", "curl", "203.0.113.10", "curl/8.0") {
		t.Fatal("expected ip+ua match")
	}
}

func TestMatchesEmailSuffix(t *testing.T) {
	t.Parallel()
	if !MatchesEmailSuffix("user@365.liout.com", "@365.liout.com") {
		t.Fatal("expected exact suffix match")
	}
	if MatchesEmailSuffix("user@mail.365.liout.com", "@365.liout.com") {
		t.Fatal("exact suffix should not match subdomain")
	}
	if !MatchesEmailSuffix("user@mail.365.liout.com", "*.365.liout.com") {
		t.Fatal("expected wildcard suffix match")
	}
	if MatchesEmailSuffix("not-an-email", "@365.liout.com") {
		t.Fatal("invalid email should not match")
	}
}

func TestValidateEmailRegexPattern(t *testing.T) {
	t.Parallel()
	if !ValidateEmailRegexPattern(`@365\.liout\.com$`) {
		t.Fatal("expected valid regex")
	}
	if ValidateEmailRegexPattern("(") {
		t.Fatal("expected invalid regex")
	}
}

func TestMatchesEmailRegex(t *testing.T) {
	t.Parallel()
	emails := []string{
		"hyi2eo8mze@365.liout.com",
		"a.drie.ncaliffhvq81@gmail.com",
		"msilntl3370+8jsuf2nntn2iz@hotmail.com",
		"cppttlf4390v+baxxsxjh9zj@hotmail.com",
	}
	rules := map[string]bool{
		`@365\.liout\.com$`:            true,
		`\.[^@]*\.[^@]+@gmail\.com$`:   true,
		`\+[^@]+@hotmail\.com$`:         true,
		`^[^@]+@example\.com$`:          false,
	}
	for pattern, wantAny := range rules {
		matched := 0
		for _, email := range emails {
			if MatchesEmailRegex(email, pattern) {
				matched++
			}
		}
		if wantAny && matched == 0 {
			t.Fatalf("pattern %q should match at least one sample email", pattern)
		}
		if !wantAny && matched > 0 {
			t.Fatalf("pattern %q should not match sample emails", pattern)
		}
	}
	if !MatchesEmailRegex("hyi2eo8mze@365.liout.com", `@365\.liout\.com$`) {
		t.Fatal("expected liout regex match")
	}
	if MatchesEmailRegex("user@example.com", `@365\.liout\.com$`) {
		t.Fatal("expected liout regex mismatch")
	}
}
