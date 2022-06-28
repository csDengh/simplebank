package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/csdengh/cur_blank/utils"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) *Account {
	args := CreateAccountParams{
		Owner:    utils.RandomOwner(),
		Balance:  utils.RandomMoney(),
		Currency: utils.RandomCurrency(),
	}

	accout, err := testQueries.CreateAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, accout)

	require.Equal(t, args.Owner, accout.Owner)
	require.Equal(t, args.Balance, accout.Balance)
	require.Equal(t, args.Currency, accout.Currency)
	require.NotZero(t, accout.ID)
	require.NotZero(t, accout.CreatedAt)
	return &accout
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	accout_expect := createRandomAccount(t)
	accout_actual, err := testQueries.GetAccount(context.Background(), accout_expect.ID)

	require.NoError(t, err)
	require.NotEmpty(t, accout_actual)

	require.Equal(t, accout_expect.ID, accout_actual.ID)
	require.Equal(t, accout_expect.Owner, accout_actual.Owner)
	require.Equal(t, accout_expect.Balance, accout_actual.Balance)
	require.Equal(t, accout_expect.Currency, accout_actual.Currency)
	require.WithinDuration(t, accout_expect.CreatedAt, accout_actual.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	accout_expect := createRandomAccount(t)

	accout_expect_parm := UpdateAccountParams{
		ID:      accout_expect.ID,
		Balance: utils.RandomMoney(),
	}

	accout_actual, err := testQueries.UpdateAccount(context.Background(), accout_expect_parm)

	require.NoError(t, err)
	require.NotEmpty(t, accout_actual)

	require.Equal(t, accout_expect_parm.ID, accout_actual.ID)
	require.Equal(t, accout_expect_parm.Balance, accout_actual.Balance)

	require.Equal(t, accout_expect.Owner, accout_actual.Owner)
	require.Equal(t, accout_expect.Currency, accout_actual.Currency)
	require.WithinDuration(t, accout_expect.CreatedAt, accout_actual.CreatedAt, time.Second)
}

func TestDeleteAccount(t *testing.T) {
	accout_expect := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), accout_expect.ID)
	require.NoError(t, err)

	accout_actual, err := testQueries.GetAccount(context.Background(), accout_expect.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, accout_actual)
}

func TestListAccount(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	accounts, err := testQueries.ListAccounts(context.Background(), ListAccountsParams{
		Limit:  5,
		Offset: 5,
	})

	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		require.NotEmpty(t, accounts[i])
	}
}
