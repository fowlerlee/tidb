package bookkeeping

import (
	"context"
	"testing"

	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookkeepingService(t *testing.T) {
	// Set up test database
	testDB, err := testutil.NewTestDB()
	require.NoError(t, err)
	defer testDB.Cleanup()

	// Initialize service
	svc := NewService(testDB.DB)

	t.Run("CreateAccount", func(t *testing.T) {
		// Test creating a valid account
		account := testutil.GenerateAccount(0)
		err := svc.CreateAccount(context.Background(), account)
		assert.NoError(t, err)
		assert.NotZero(t, account.ID)

		// Test creating duplicate account code
		dupAccount := testutil.GenerateAccount(0)
		dupAccount.AccountCode = account.AccountCode
		err = svc.CreateAccount(context.Background(), dupAccount)
		assert.Error(t, err)

		// Test creating account with invalid type
		invalidAccount := testutil.GenerateAccount(0)
		invalidAccount.Type = "InvalidType"
		err = svc.CreateAccount(context.Background(), invalidAccount)
		assert.Error(t, err)
	})

	t.Run("CreateJournalEntry", func(t *testing.T) {
		// Create test accounts first
		debitAccount := testutil.GenerateAccount(0)
		err := svc.CreateAccount(context.Background(), debitAccount)
		require.NoError(t, err)

		creditAccount := testutil.GenerateAccount(0)
		err = svc.CreateAccount(context.Background(), creditAccount)
		require.NoError(t, err)

		// Test creating a valid journal entry with balanced debits and credits
		entry := testutil.GenerateJournalEntry(0)
		lines := []models.JournalLine{
			{
				AccountID:   debitAccount.ID,
				Description: "Test debit",
				DebitAmount: 1000.0,
			},
			{
				AccountID:    creditAccount.ID,
				Description:  "Test credit",
				CreditAmount: 1000.0,
			},
		}

		err = svc.CreateJournalEntry(context.Background(), entry, lines)
		assert.NoError(t, err)
		assert.NotZero(t, entry.ID)

		// Test creating entry with unbalanced debits and credits
		unbalancedEntry := testutil.GenerateJournalEntry(0)
		unbalancedLines := []models.JournalLine{
			{
				AccountID:   debitAccount.ID,
				Description: "Test debit",
				DebitAmount: 1000.0,
			},
			{
				AccountID:    creditAccount.ID,
				Description:  "Test credit",
				CreditAmount: 900.0, // Unbalanced amount
			},
		}

		err = svc.CreateJournalEntry(context.Background(), unbalancedEntry, unbalancedLines)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "debits must equal credits")

		// Test creating entry with non-existent account
		invalidEntry := testutil.GenerateJournalEntry(0)
		invalidLines := []models.JournalLine{
			{
				AccountID:   999999, // Non-existent account
				Description: "Test debit",
				DebitAmount: 1000.0,
			},
			{
				AccountID:    creditAccount.ID,
				Description:  "Test credit",
				CreditAmount: 1000.0,
			},
		}

		err = svc.CreateJournalEntry(context.Background(), invalidEntry, invalidLines)
		assert.Error(t, err)
	})

	t.Run("GetJournalEntry", func(t *testing.T) {
		// Create test data
		debitAccount := testutil.GenerateAccount(0)
		err := svc.CreateAccount(context.Background(), debitAccount)
		require.NoError(t, err)

		creditAccount := testutil.GenerateAccount(0)
		err = svc.CreateAccount(context.Background(), creditAccount)
		require.NoError(t, err)

		entry := testutil.GenerateJournalEntry(0)
		lines := []models.JournalLine{
			{
				AccountID:   debitAccount.ID,
				Description: "Test debit",
				DebitAmount: 1000.0,
			},
			{
				AccountID:    creditAccount.ID,
				Description:  "Test credit",
				CreditAmount: 1000.0,
			},
		}

		err = svc.CreateJournalEntry(context.Background(), entry, lines)
		require.NoError(t, err)

		// Test retrieving existing entry
		fetchedEntry, fetchedLines, err := svc.GetJournalEntry(context.Background(), entry.ID)
		assert.NoError(t, err)
		assert.NotNil(t, fetchedEntry)
		assert.Equal(t, entry.EntryNumber, fetchedEntry.EntryNumber)
		assert.Len(t, fetchedLines, 2)
		assert.Equal(t, fetchedLines[0].DebitAmount, 1000.0)
		assert.Equal(t, fetchedLines[1].CreditAmount, 1000.0)

		// Test retrieving non-existent entry
		fetchedEntry, fetchedLines, err = svc.GetJournalEntry(context.Background(), 999999)
		assert.Error(t, err)
		assert.Nil(t, fetchedEntry)
		assert.Nil(t, fetchedLines)
	})

	t.Run("GetAccountBalance", func(t *testing.T) {
		// Create test account
		account := testutil.GenerateAccount(0)
		err := svc.CreateAccount(context.Background(), account)
		require.NoError(t, err)

		// Create a journal entry affecting the account
		entry := testutil.GenerateJournalEntry(0)
		lines := []models.JournalLine{
			{
				AccountID:   account.ID,
				Description: "Test debit",
				DebitAmount: 1000.0,
			},
			{
				AccountID:    account.ID, // Same account for testing
				Description:  "Test credit",
				CreditAmount: 400.0,
			},
		}

		err = svc.CreateJournalEntry(context.Background(), entry, lines)
		require.NoError(t, err)

		// Test getting account balance
		balance, err := svc.GetAccountBalance(context.Background(), account.ID)
		assert.NoError(t, err)
		assert.Equal(t, 600.0, balance) // 1000 debit - 400 credit = 600 net debit

		// Test getting balance for non-existent account
		balance, err = svc.GetAccountBalance(context.Background(), 999999)
		assert.Error(t, err)
		assert.Zero(t, balance)
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		// Create test account
		account := testutil.GenerateAccount(0)
		err := svc.CreateAccount(context.Background(), account)
		require.NoError(t, err)

		// Get initial balance
		initialBalance, err := svc.GetAccountBalance(context.Background(), account.ID)
		require.NoError(t, err)

		// Try to create an invalid journal entry
		entry := testutil.GenerateJournalEntry(0)
		lines := []models.JournalLine{
			{
				AccountID:   account.ID,
				Description: "Test debit",
				DebitAmount: 1000.0,
			},
			{
				AccountID:    999999, // Non-existent account
				Description:  "Test credit",
				CreditAmount: 1000.0,
			},
		}

		err = svc.CreateJournalEntry(context.Background(), entry, lines)
		assert.Error(t, err)

		// Verify account balance hasn't changed
		currentBalance, err := svc.GetAccountBalance(context.Background(), account.ID)
		assert.NoError(t, err)
		assert.Equal(t, initialBalance, currentBalance)
	})

	t.Run("ConcurrentTransactions", func(t *testing.T) {
		// Create test account
		account := testutil.GenerateAccount(0)
		err := svc.CreateAccount(context.Background(), account)
		require.NoError(t, err)

		// Create multiple journal entries concurrently
		done := make(chan bool)
		for i := 0; i < 5; i++ {
			go func() {
				entry := testutil.GenerateJournalEntry(0)
				lines := []models.JournalLine{
					{
						AccountID:   account.ID,
						Description: "Test debit",
						DebitAmount: 100.0,
					},
					{
						AccountID:    account.ID,
						Description:  "Test credit",
						CreditAmount: 100.0,
					},
				}
				err := svc.CreateJournalEntry(context.Background(), entry, lines)
				assert.NoError(t, err)
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-done
		}
	})
}
