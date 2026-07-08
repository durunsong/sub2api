// Package accessban implements matching helpers for global access ban rules.
package accessban

import (
	"regexp"
	"strings"
	"sync"

	ippkg "github.com/Wei-Shaw/sub2api/internal/pkg/ip"
)

const (
	RuleTypeIP          = "ip"
	RuleTypeUA          = "ua"
	RuleTypeIPUA        = "ip_ua"
	RuleTypeEmailSuffix = "email_suffix"
	RuleTypeEmailRegex  = "email_regex"
)

var validRuleTypes = map[string]struct{}{
	RuleTypeIP:          {},
	RuleTypeUA:          {},
	RuleTypeIPUA:        {},
	RuleTypeEmailSuffix: {},
	RuleTypeEmailRegex:  {},
}

var emailRegexCache sync.Map

// NormalizeRuleType returns a supported rule type or empty string.
func NormalizeRuleType(raw string) string {
	t := strings.ToLower(strings.TrimSpace(raw))
	if _, ok := validRuleTypes[t]; ok {
		return t
	}
	return ""
}

// ValidateRuleType reports whether raw is a supported rule type.
func ValidateRuleType(raw string) bool {
	return NormalizeRuleType(raw) != ""
}

// ValidateUAPattern validates UA match text.
func ValidateUAPattern(raw string) bool {
	value := strings.TrimSpace(raw)
	return len(value) >= 2 && len(value) <= 255 && !strings.ContainsAny(value, "\r\n")
}

// MatchesClient checks whether a client-side ban rule matches IP and/or UA.
func MatchesClient(ruleType, pattern, uaPattern, clientIP, userAgent string) bool {
	switch NormalizeRuleType(ruleType) {
	case RuleTypeIP:
		return ippkg.MatchesPattern(clientIP, pattern)
	case RuleTypeUA:
		return matchesUAPattern(userAgent, pattern)
	case RuleTypeIPUA:
		return ippkg.MatchesPattern(clientIP, pattern) && matchesUAPattern(userAgent, uaPattern)
	default:
		return false
	}
}

// MatchesEmailSuffix checks whether an email matches a suffix ban pattern.
// Pattern must already be normalized (@domain or *.domain).
func MatchesEmailSuffix(email, pattern string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	if pattern == "" {
		return false
	}
	_, domain, ok := splitEmail(email)
	if !ok {
		return false
	}
	suffix := "@" + domain
	if strings.HasPrefix(pattern, "*.") {
		return domainMatchesWildcard(domain, pattern)
	}
	if strings.HasPrefix(pattern, "@") {
		return suffix == pattern
	}
	return false
}

// ValidateEmailRegexPattern validates a regex pattern for email ban rules.
func ValidateEmailRegexPattern(raw string) bool {
	pattern := strings.TrimSpace(raw)
	if len(pattern) < 3 || len(pattern) > 255 || strings.ContainsAny(pattern, "\r\n") {
		return false
	}
	_, err := compileEmailRegex(pattern)
	return err == nil
}

// MatchesEmailRegex checks whether a full email matches a regex ban pattern.
// Matching uses lowercase trimmed email as input.
func MatchesEmailRegex(email, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return false
	}
	re, err := compileEmailRegex(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(normalized)
}

func compileEmailRegex(pattern string) (*regexp.Regexp, error) {
	if cached, ok := emailRegexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp), nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	emailRegexCache.Store(pattern, re)
	return re, nil
}

func matchesUAPattern(userAgent, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	if ua == "" {
		return false
	}
	return strings.Contains(ua, strings.ToLower(pattern))
}

func splitEmail(raw string) (local, domain string, ok bool) {
	email := strings.ToLower(strings.TrimSpace(raw))
	local, domain, found := strings.Cut(email, "@")
	if !found || local == "" || domain == "" || strings.Contains(domain, "@") {
		return "", "", false
	}
	return local, domain, true
}

func domainMatchesWildcard(domain, allowed string) bool {
	base := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(allowed)), "*.")
	if base == "" || !isValidEmailDomain(base) {
		return false
	}
	return domain == base || strings.HasSuffix(domain, "."+base)
}

func isValidEmailDomain(domain string) bool {
	if domain == "" || strings.Contains(domain, "@") {
		return false
	}
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" || len(part) > 63 {
			return false
		}
		for _, r := range part {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				continue
			}
			return false
		}
		if part[0] == '-' || part[len(part)-1] == '-' {
			return false
		}
	}
	return true
}
