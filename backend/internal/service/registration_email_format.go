package service

import (
	"net/mail"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrInvalidEmailFormat = infraerrors.BadRequest("INVALID_EMAIL_FORMAT", "invalid email format")

// ValidateRegistrationEmailFormat applies stricter email sanity checks than HTML5 email input.
func ValidateRegistrationEmailFormat(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrInvalidEmailFormat
	}
	if len(email) > 254 {
		return ErrInvalidEmailFormat
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmailFormat
	}
	local, domain, ok := splitEmailForPolicy(email)
	if !ok || !isValidRegistrationEmailDomain(domain) {
		return ErrInvalidEmailFormat
	}
	if len(local) > 64 || strings.Contains(local, "..") || strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
		return ErrInvalidEmailFormat
	}
	return nil
}
