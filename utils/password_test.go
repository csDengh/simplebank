package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashpwd(t *testing.T) {
	pwd := RandomString(6)

	hashpwd, err := HashPassword(pwd)
	require.NoError(t, err)

	err = ConfirmPwd(hashpwd, pwd)
	require.NoError(t, err)

	pwd1 := RandomString(6)

	err = ConfirmPwd(hashpwd, pwd1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
}
