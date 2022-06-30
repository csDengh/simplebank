package db

import (
	"context"
	"testing"
	"time"

	"github.com/csdengh/cur_blank/utils"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) *User {
	pwd := utils.RandomString(6)
	hashpwd, err := utils.HashPassword(pwd)
	if err != nil {
		return &User{}
	}

	args := CreateUserParams{
		Username:       utils.RandomOwner(),
		HashedPassword: hashpwd,
		FullName:       utils.RandomOwner(),
		Email:          utils.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, args.Username, user.Username)
	require.Equal(t, args.HashedPassword, user.HashedPassword)
	require.Equal(t, args.FullName, user.FullName)
	require.Equal(t, args.Email, user.Email)

	require.NotZero(t, user.CreatedAt)
	return &user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	accout_expect := createRandomUser(t)
	accout_actual, err := testQueries.GetUser(context.Background(), accout_expect.Username)

	require.NoError(t, err)
	require.NotEmpty(t, accout_actual)

	require.Equal(t, accout_expect.Username, accout_actual.Username)
	require.Equal(t, accout_expect.HashedPassword, accout_actual.HashedPassword)
	require.Equal(t, accout_expect.FullName, accout_actual.FullName)
	require.Equal(t, accout_expect.Email, accout_actual.Email)
	require.WithinDuration(t, accout_expect.CreatedAt, accout_actual.CreatedAt, time.Second)
}
