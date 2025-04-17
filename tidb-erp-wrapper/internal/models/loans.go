package models

import "time"

// Loan represents a loan given by a company to another company
type Loan struct {
	ID                int64     `json:"id"`
	LoanNumber        string    `json:"loan_number"`
	LenderCompanyID   int64     `json:"lender_company_id"`
	BorrowerCompanyID int64     `json:"borrower_company_id"`
	PrincipalAmount   float64   `json:"principal_amount"`
	OutstandingAmount float64   `json:"outstanding_amount"`
	InterestRate      float64   `json:"interest_rate"` // Annual interest rate as percentage
	StartDate         time.Time `json:"start_date"`
	MaturityDate      time.Time `json:"maturity_date"`
	PaymentFrequency  string    `json:"payment_frequency"` // Monthly, Quarterly, Semi-Annually, Annually
	PaymentDay        int       `json:"payment_day"`       // Day of the period when payment is due
	Status            string    `json:"status"`            // Active, Completed, Defaulted, Pending
	CollateralDetails string    `json:"collateral_details,omitempty"`
	Notes             string    `json:"notes,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// LoanPayment represents a payment made against a loan
type LoanPayment struct {
	ID                   int64     `json:"id"`
	LoanID               int64     `json:"loan_id"`
	PaymentNumber        string    `json:"payment_number"`
	PaymentDate          time.Time `json:"payment_date"`
	PrincipalAmount      float64   `json:"principal_amount"`
	InterestAmount       float64   `json:"interest_amount"`
	TotalAmount          float64   `json:"total_amount"`
	PaymentMethod        string    `json:"payment_method"`
	TransactionReference string    `json:"transaction_reference,omitempty"`
	Notes                string    `json:"notes,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// LoanScheduleItem represents a scheduled payment for a loan
type LoanScheduleItem struct {
	ID              int64     `json:"id"`
	LoanID          int64     `json:"loan_id"`
	ScheduledDate   time.Time `json:"scheduled_date"`
	PrincipalAmount float64   `json:"principal_amount"`
	InterestAmount  float64   `json:"interest_amount"`
	TotalAmount     float64   `json:"total_amount"`
	Status          string    `json:"status"` // Pending, Paid, Overdue
	ActualPaymentID *int64    `json:"actual_payment_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Company represents either a lender or borrower company
type Company struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	RegistrationNumber string    `json:"registration_number"`
	TaxID              string    `json:"tax_id"`
	ContactPerson      string    `json:"contact_person"`
	Email              string    `json:"email"`
	Phone              string    `json:"phone"`
	Address            string    `json:"address"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
