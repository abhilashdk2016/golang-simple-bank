package db

import (
	"context"
	"testing"

	"github.com/abhilashdk2016/golang-simple-bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomeTransfer(t *testing.T) Transfer {
	arg := CreateTransferParams{
		FromAccountID: 5,
		ToAccountID:   6,
		Amount:        util.RandomMoney(),
	}

	entry, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)
	require.Equal(t, arg.FromAccountID, entry.FromAccountID)
	require.Equal(t, arg.ToAccountID, entry.ToAccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	return entry
}

func TestCreateTransfer(t *testing.T) {
	createRandomeTransfer(t)
}

func TestGetTransfer(t *testing.T) {
	transfer := createRandomeTransfer(t)
	transfer1, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer1)

	require.Equal(t, transfer.FromAccountID, transfer1.FromAccountID)
	require.Equal(t, transfer.ToAccountID, transfer1.ToAccountID)
	require.Equal(t, transfer.Amount, transfer1.Amount)
}

func TestListTransfers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomeTransfer(t)
	}
	arg := ListTransfersParams{
		FromAccountID: 5,
		ToAccountID:   6,
		Limit:         5,
		Offset:        5,
	}
	transfers, err := testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}
