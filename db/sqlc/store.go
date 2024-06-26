package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
	Querier
}

type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)

	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %w, roll back error: %w", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account_id"`
	ToAccount   Account  `json:"to_account_id"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

//var txKey = struct{}{}

func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		//txName := ctx.Value(txKey)
		//fmt.Println(txName, "create transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		//fmt.Println(txName, "create entry1")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		//fmt.Println(txName, "create entry2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		//fmt.Println(txName, "get account 1")
		// get account -> update balance
		// account1, err := q.GetAccountForUpdates(ctx, arg.FromAccountID)
		// if err != nil {
		// 	return err
		// }

		//fmt.Println(txName, "update account 1")
		if arg.FromAccountID < arg.ToAccountID {
			err = q.AddUpdateAccount(ctx, AddUpdateAccountParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err != nil {
				return err
			}
			result.FromAccount, err = q.GetAccount(ctx, arg.FromAccountID)
			if err != nil {
				return err
			}

			//fmt.Println(txName, "get account 2")
			// account2, err := q.GetAccountForUpdates(ctx, arg.ToAccountID)
			// if err != nil {
			// 	return err
			// }

			//fmt.Println(txName, "update account 2")
			err = q.AddUpdateAccount(ctx, AddUpdateAccountParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}
			result.ToAccount, err = q.GetAccount(ctx, arg.ToAccountID)
			if err != nil {
				return err
			}
		} else {
			err = q.AddUpdateAccount(ctx, AddUpdateAccountParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}
			result.ToAccount, err = q.GetAccount(ctx, arg.ToAccountID)
			if err != nil {
				return err
			}

			err = q.AddUpdateAccount(ctx, AddUpdateAccountParams{
				ID:     arg.FromAccountID,
				Amount: -arg.Amount,
			})
			if err != nil {
				return err
			}
			result.FromAccount, err = q.GetAccount(ctx, arg.FromAccountID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}
