package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/andreanpradanaa/simple-bank-app/util"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	args := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotNil(t, account)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	require.Equal(t, args.Owner, account.Owner)
	require.Equal(t, args.Balance, account.Balance)
	require.Equal(t, args.Currency, account.Currency)

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)

}

func TestDeleteAccount(t *testing.T) {
	account := createRandomAccount(t)
	err := testQueries.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)

	getAccount, err := testQueries.GetAccount(context.Background(), account.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, getAccount)
}

func TestGetAccount(t *testing.T) {
	account := createRandomAccount(t)
	testGetAccount, err := testQueries.GetAccount(context.Background(), account.ID)

	require.NoError(t, err)
	require.NotNil(t, testGetAccount)

	require.Equal(t, account.ID, testGetAccount.ID)
	require.Equal(t, account.Owner, testGetAccount.Owner)
	require.Equal(t, account.Balance, testGetAccount.Balance)
	require.Equal(t, account.Currency, testGetAccount.Currency)
	require.Equal(t, account.CreatedAt, testGetAccount.CreatedAt)
	// require.WithinDuration(t, account.CreatedAt, actualAccount.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	account := createRandomAccount(t)

	args := UpdateAccountParams{
		ID:      account.ID,
		Balance: 100,
	}

	updatedAccount, err := testQueries.UpdateAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotNil(t, updatedAccount)

	require.Equal(t, account.ID, updatedAccount.ID)
	require.Equal(t, account.Owner, updatedAccount.Owner)
	require.Equal(t, args.Balance, updatedAccount.Balance)
	require.Equal(t, account.Currency, updatedAccount.Currency)
	require.Equal(t, account.CreatedAt, updatedAccount.CreatedAt)
}

func TestListAccount(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	listAccounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, listAccounts, 5)

	for _, account := range listAccounts {
		require.NotEmpty(t, account)
	}
}
