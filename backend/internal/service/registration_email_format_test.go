package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRegistrationEmailFormat(t *testing.T) {
	require.NoError(t, ValidateRegistrationEmailFormat("user@example.com"))
	require.NoError(t, ValidateRegistrationEmailFormat("user.name+tag@365.liout.com"))
	require.ErrorIs(t, ValidateRegistrationEmailFormat(""), ErrInvalidEmailFormat)
	require.ErrorIs(t, ValidateRegistrationEmailFormat("bad@"), ErrInvalidEmailFormat)
	require.ErrorIs(t, ValidateRegistrationEmailFormat(".user@example.com"), ErrInvalidEmailFormat)
	require.ErrorIs(t, ValidateRegistrationEmailFormat("user..name@example.com"), ErrInvalidEmailFormat)
	require.ErrorIs(t, ValidateRegistrationEmailFormat("user@invalid_domain"), ErrInvalidEmailFormat)
}
