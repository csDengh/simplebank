package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, req *TransferTxParams) *TransferTxResult
}

type SqlStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SqlStore{
		Queries: New(db),
		db:      db,
	}
}

func (s *SqlStore) execTran(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx)

	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("q error=%s rb error=%s", err, rbErr)
		}
		return err
	}
	tx.Commit()
	return nil
}

type TransferTxParams struct {
	FromAccountId int64 `json:"from_account_id"`
	ToAccountId   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
	err         error
}

func (s *SqlStore) TransferTx(ctx context.Context, req *TransferTxParams) *TransferTxResult {
	var result TransferTxResult

	err := s.execTran(ctx, func(q *Queries) error {
		var err error

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccoutID: req.FromAccountId,
			Amount:   -req.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccoutID: req.ToAccountId,
			Amount:   req.Amount,
		})
		if err != nil {
			return err
		}

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: req.FromAccountId,
			ToAccountID:   req.ToAccountId,
			Amount:        req.Amount,
		})
		if err != nil {
			return err
		}

		if req.FromAccountId < req.ToAccountId {
			result.FromAccount, result.ToAccount, err = updateAccounts(ctx, q, &UpdateAccountBalanceParams{Amount: -req.Amount, ID: req.FromAccountId}, &UpdateAccountBalanceParams{Amount: req.Amount, ID: req.ToAccountId})
		} else {
			result.ToAccount, result.FromAccount, err = updateAccounts(ctx, q, &UpdateAccountBalanceParams{Amount: req.Amount, ID: req.ToAccountId}, &UpdateAccountBalanceParams{Amount: -req.Amount, ID: req.FromAccountId})
		}

		return err
	})
	result.err = err
	return &result
}

func updateAccounts(ctx context.Context, q *Queries, lowAccount, highAccount *UpdateAccountBalanceParams) (account1 Account, account2 Account, err error) {
	account1, err = q.UpdateAccountBalance(ctx, *lowAccount)
	if err != nil {
		return
	}
	account2, err = q.UpdateAccountBalance(ctx, *highAccount)
	if err != nil {
		return
	}

	return
}
