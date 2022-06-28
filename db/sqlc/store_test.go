package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranfer(t *testing.T) {
	s := NewStore(testDb)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 5
	amount := int64(10)

	req := &TransferTxParams{
		FromAccountId: account1.ID,
		ToAccountId:   account2.ID,
		Amount:        amount,
	}

	resChan := make(chan *TransferTxResult)

	for i := 0; i < n; i++ {

		go func() {
			res := s.TransferTx(context.Background(), req)
			resChan <- res
		}()

	}

	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		res := <-resChan
		require.NoError(t, res.err)

		transfer := res.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err := s.GetTransfer(context.Background(), res.Transfer.ID)
		require.NoError(t, err)

		fromEntry := res.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccoutID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = s.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := res.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccoutID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = s.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		fromAccount := res.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := res.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // 1 * amount, 2 * amount, 3 * amount, ..., n * amount

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	account1Actual, err := s.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account1Actual)

	account2Actual, err := s.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account2Actual)

	require.Equal(t, account1.Balance-int64(n)*amount, account1Actual.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, account2Actual.Balance)
}

func TestTranferDeadLock(t *testing.T) {
	s := NewStore(testDb)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 10
	amount := int64(10)

	resChan := make(chan *TransferTxResult)

	for i := 0; i < n; i++ {

		fromId := account1.ID
		toId := account2.ID

		if i%2 == 1 {
			fromId = account2.ID
			toId = account1.ID
		}

		req := &TransferTxParams{
			FromAccountId: fromId,
			ToAccountId:   toId,
			Amount:        amount,
		}

		go func() {
			res := s.TransferTx(context.Background(), req)
			resChan <- res
		}()

	}

	for i := 0; i < n; i++ {
		res := <-resChan
		require.NoError(t, res.err)
	}

	account1Actual, err := s.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account1Actual)

	account2Actual, err := s.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account2Actual)

	require.Equal(t, account1.Balance, account1Actual.Balance)
	require.Equal(t, account2.Balance, account2Actual.Balance)
}
