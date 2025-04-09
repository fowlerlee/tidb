package bookkeeping

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tidb-erp-wrapper/internal/db"
	"tidb-erp-wrapper/internal/models"
)

type Service struct {
	db *db.DBHandler
}

func NewService(db *db.DBHandler) *Service {
	return &Service{db: db}
}

func (s *Service) CreateJournalEntry(ctx context.Context, entry *models.JournalEntry, lines []models.JournalLine) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert journal entry
	query := `
		INSERT INTO journal_entries (
			entry_number, date, reference, description,
			status, posted_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		entry.EntryNumber,
		entry.Date,
		entry.Reference,
		entry.Description,
		entry.Status,
		entry.PostedBy,
	)
	if err != nil {
		return fmt.Errorf("error creating journal entry: %v", err)
	}

	journalID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting journal entry ID: %v", err)
	}
	entry.ID = journalID

	// Validate debits equal credits
	var totalDebits, totalCredits float64
	for _, line := range lines {
		totalDebits += line.DebitAmount
		totalCredits += line.CreditAmount
	}
	if totalDebits != totalCredits {
		return errors.New("debits must equal credits")
	}

	// Insert journal lines
	for _, line := range lines {
		query := `
			INSERT INTO journal_lines (
				journal_id, account_id, description,
				debit_amount, credit_amount, currency_code,
				exchange_rate, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		`
		_, err := tx.ExecContext(ctx, query,
			journalID,
			line.AccountID,
			line.Description,
			line.DebitAmount,
			line.CreditAmount,
			line.CurrencyCode,
			line.ExchangeRate,
		)
		if err != nil {
			return fmt.Errorf("error creating journal line: %v", err)
		}

		// Update account balances
		if line.DebitAmount > 0 {
			if err := s.updateAccountBalance(ctx, tx, line.AccountID, line.DebitAmount, true); err != nil {
				return err
			}
		}
		if line.CreditAmount > 0 {
			if err := s.updateAccountBalance(ctx, tx, line.AccountID, line.CreditAmount, false); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *Service) updateAccountBalance(ctx context.Context, tx *sql.Tx, accountID int64, amount float64, isDebit bool) error {
	var query string
	if isDebit {
		query = "UPDATE accounts SET balance = balance + ? WHERE id = ?"
	} else {
		query = "UPDATE accounts SET balance = balance - ? WHERE id = ?"
	}

	result, err := tx.ExecContext(ctx, query, amount, accountID)
	if err != nil {
		return fmt.Errorf("error updating account balance: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if rows != 1 {
		return fmt.Errorf("account with ID %d not found", accountID)
	}

	return nil
}

func (s *Service) CreateAccount(ctx context.Context, account *models.Account) error {
	query := `
		INSERT INTO accounts (
			account_code, name, type, sub_type,
			description, balance, currency_code,
			is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		account.AccountCode,
		account.Name,
		account.Type,
		account.SubType,
		account.Description,
		account.Balance,
		account.CurrencyCode,
		account.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating account: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	account.ID = id
	return nil
}

func (s *Service) GetAccountBalance(ctx context.Context, accountID int64) (float64, error) {
	var balance float64
	query := "SELECT balance FROM accounts WHERE id = ?"
	err := s.db.DB().QueryRowContext(ctx, query, accountID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("account not found")
		}
		return 0, fmt.Errorf("error querying account balance: %v", err)
	}
	return balance, nil
}

func (s *Service) GetJournalEntry(ctx context.Context, id int64) (*models.JournalEntry, []models.JournalLine, error) {
	entry := &models.JournalEntry{}
	query := `
		SELECT id, entry_number, date, reference,
			   description, status, posted_by,
			   created_at, updated_at
		FROM journal_entries
		WHERE id = ?
	`
	err := s.db.DB().QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.EntryNumber,
		&entry.Date,
		&entry.Reference,
		&entry.Description,
		&entry.Status,
		&entry.PostedBy,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("journal entry not found")
		}
		return nil, nil, fmt.Errorf("error querying journal entry: %v", err)
	}

	lines := []models.JournalLine{}
	query = `
		SELECT id, account_id, description,
			   debit_amount, credit_amount,
			   currency_code, exchange_rate,
			   created_at, updated_at
		FROM journal_lines
		WHERE journal_id = ?
	`
	rows, err := s.db.DB().QueryContext(ctx, query, id)
	if err != nil {
		return nil, nil, fmt.Errorf("error querying journal lines: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var line models.JournalLine
		err := rows.Scan(
			&line.ID,
			&line.AccountID,
			&line.Description,
			&line.DebitAmount,
			&line.CreditAmount,
			&line.CurrencyCode,
			&line.ExchangeRate,
			&line.CreatedAt,
			&line.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning journal line: %v", err)
		}
		line.JournalID = id
		lines = append(lines, line)
	}

	return entry, lines, nil
}
