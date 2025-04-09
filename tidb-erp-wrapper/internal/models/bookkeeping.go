package models

import "time"

type Account struct {
	ID           int64     `json:"id"`
	AccountCode  string    `json:"account_code"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // Asset, Liability, Equity, Revenue, Expense
	SubType      string    `json:"sub_type"`
	Description  string    `json:"description"`
	Balance      float64   `json:"balance"`
	CurrencyCode string    `json:"currency_code"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type JournalEntry struct {
	ID          int64     `json:"id"`
	EntryNumber string    `json:"entry_number"`
	Date        time.Time `json:"date"`
	Reference   string    `json:"reference"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	PostedBy    string    `json:"posted_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type JournalLine struct {
	ID           int64     `json:"id"`
	JournalID    int64     `json:"journal_id"`
	AccountID    int64     `json:"account_id"`
	Description  string    `json:"description"`
	DebitAmount  float64   `json:"debit_amount"`
	CreditAmount float64   `json:"credit_amount"`
	CurrencyCode string    `json:"currency_code"`
	ExchangeRate float64   `json:"exchange_rate"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type FiscalPeriod struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	IsClosed  bool      `json:"is_closed"`
	ClosedBy  string    `json:"closed_by"`
	ClosedAt  time.Time `json:"closed_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
