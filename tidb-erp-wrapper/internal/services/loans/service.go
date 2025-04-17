package loans

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/db"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
)

// Service handles the business logic for loan management
type Service struct {
	db *db.DBHandler
}

// NewService creates a new instance of the loans service
func NewService(db *db.DBHandler) *Service {
	return &Service{db: db}
}

// CreateCompany creates a new company record (lender or borrower)
func (s *Service) CreateCompany(ctx context.Context, company *models.Company) error {
	query := `
		INSERT INTO companies (
			name, registration_number, tax_id, contact_person,
			email, phone, address, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		company.Name,
		company.RegistrationNumber,
		company.TaxID,
		company.ContactPerson,
		company.Email,
		company.Phone,
		company.Address,
		company.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating company: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	company.ID = id
	return nil
}

// GetCompany retrieves a company by ID
func (s *Service) GetCompany(ctx context.Context, id int64) (*models.Company, error) {
	query := `
		SELECT id, name, registration_number, tax_id, contact_person,
			   email, phone, address, is_active, created_at, updated_at
		FROM companies
		WHERE id = ?
	`
	company := &models.Company{}
	err := s.db.DB().QueryRowContext(ctx, query, id).Scan(
		&company.ID,
		&company.Name,
		&company.RegistrationNumber,
		&company.TaxID,
		&company.ContactPerson,
		&company.Email,
		&company.Phone,
		&company.Address,
		&company.IsActive,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("company not found")
		}
		return nil, fmt.Errorf("error querying company: %v", err)
	}
	return company, nil
}

// GetCompanyByRegistrationNumber retrieves a company by registration number
func (s *Service) GetCompanyByRegistrationNumber(ctx context.Context, regNum string) (*models.Company, error) {
	query := `
		SELECT id, name, registration_number, tax_id, contact_person,
			   email, phone, address, is_active, created_at, updated_at
		FROM companies
		WHERE registration_number = ?
	`
	company := &models.Company{}
	err := s.db.DB().QueryRowContext(ctx, query, regNum).Scan(
		&company.ID,
		&company.Name,
		&company.RegistrationNumber,
		&company.TaxID,
		&company.ContactPerson,
		&company.Email,
		&company.Phone,
		&company.Address,
		&company.IsActive,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("company not found")
		}
		return nil, fmt.Errorf("error querying company: %v", err)
	}
	return company, nil
}

// CreateLoan creates a new loan and generates its payment schedule
func (s *Service) CreateLoan(ctx context.Context, loan *models.Loan) error {
	// Start transaction
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Validate loan data
	if loan.LenderCompanyID == loan.BorrowerCompanyID {
		return errors.New("lender and borrower cannot be the same company")
	}

	if loan.PrincipalAmount <= 0 {
		return errors.New("principal amount must be greater than zero")
	}

	if loan.InterestRate < 0 {
		return errors.New("interest rate cannot be negative")
	}

	if loan.StartDate.After(loan.MaturityDate) {
		return errors.New("start date must be before maturity date")
	}

	// Set initial outstanding amount to principal amount
	loan.OutstandingAmount = loan.PrincipalAmount

	// Insert loan
	query := `
		INSERT INTO loans (
			loan_number, lender_company_id, borrower_company_id,
			principal_amount, outstanding_amount, interest_rate,
			start_date, maturity_date, payment_frequency, payment_day,
			status, collateral_details, notes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		loan.LoanNumber,
		loan.LenderCompanyID,
		loan.BorrowerCompanyID,
		loan.PrincipalAmount,
		loan.OutstandingAmount,
		loan.InterestRate,
		loan.StartDate,
		loan.MaturityDate,
		loan.PaymentFrequency,
		loan.PaymentDay,
		loan.Status,
		loan.CollateralDetails,
		loan.Notes,
	)
	if err != nil {
		return fmt.Errorf("error creating loan: %v", err)
	}

	loanID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting loan ID: %v", err)
	}
	loan.ID = loanID

	// Generate payment schedule
	schedule, err := s.generateLoanSchedule(loan)
	if err != nil {
		return fmt.Errorf("error generating loan schedule: %v", err)
	}

	// Insert schedule items
	for _, item := range schedule {
		item.LoanID = loanID
		item.Status = "Pending"

		query := `
			INSERT INTO loan_schedule_items (
				loan_id, scheduled_date, principal_amount,
				interest_amount, total_amount, status,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		`
		_, err := tx.ExecContext(ctx, query,
			item.LoanID,
			item.ScheduledDate,
			item.PrincipalAmount,
			item.InterestAmount,
			item.TotalAmount,
			item.Status,
		)
		if err != nil {
			return fmt.Errorf("error creating loan schedule item: %v", err)
		}
	}

	return tx.Commit()
}

// generateLoanSchedule creates a payment schedule for the given loan
func (s *Service) generateLoanSchedule(loan *models.Loan) ([]models.LoanScheduleItem, error) {
	var scheduleItems []models.LoanScheduleItem

	term := s.calculateTermInMonths(loan.StartDate, loan.MaturityDate)
	if term <= 0 {
		return nil, errors.New("invalid loan term")
	}

	// Calculate payment frequency in months
	var frequencyMonths int
	switch loan.PaymentFrequency {
	case "Monthly":
		frequencyMonths = 1
	case "Quarterly":
		frequencyMonths = 3
	case "Semi-Annually":
		frequencyMonths = 6
	case "Annually":
		frequencyMonths = 12
	default:
		return nil, fmt.Errorf("unsupported payment frequency: %s", loan.PaymentFrequency)
	}

	// Calculate number of payments
	numPayments := term / frequencyMonths
	if term%frequencyMonths != 0 {
		// Round up if the term is not a multiple of the frequency
		numPayments++
	}

	// Calculate monthly interest rate
	monthlyRate := loan.InterestRate / 100 / 12

	// Use amortization formula to calculate fixed payment amount
	// P = (r*PV) / (1 - (1+r)^-n)
	// Where:
	// P = payment amount
	// r = periodic interest rate
	// PV = present value (principal)
	// n = total number of payments
	paymentAmount := (monthlyRate * loan.PrincipalAmount) /
		(1 - math.Pow(1+monthlyRate, -float64(term)))

	// Adjust payment amount for the payment frequency
	periodicPayment := paymentAmount * float64(frequencyMonths)

	remainingPrincipal := loan.PrincipalAmount
	currentDate := loan.StartDate

	for i := 0; i < numPayments && remainingPrincipal > 0; i++ {
		// Calculate next payment date
		paymentDate := s.addMonths(currentDate, frequencyMonths)

		// Adjust day of month if specified
		if loan.PaymentDay > 0 {
			paymentDate = s.setDayOfMonth(paymentDate, loan.PaymentDay)
		}

		// For the last payment, use the maturity date
		if i == numPayments-1 || paymentDate.After(loan.MaturityDate) {
			paymentDate = loan.MaturityDate
		}

		// Calculate interest for this period
		monthsInPeriod := s.calculateTermInMonths(currentDate, paymentDate)
		interestForPeriod := remainingPrincipal * monthlyRate * float64(monthsInPeriod)

		// Calculate principal portion for this payment
		var principalPayment float64
		if i == numPayments-1 {
			// Last payment - pay all remaining principal
			principalPayment = remainingPrincipal
			// Recalculate interest to be more accurate
			interestForPeriod = remainingPrincipal * monthlyRate * float64(monthsInPeriod)
		} else {
			principalPayment = periodicPayment - interestForPeriod
			// Ensure we don't pay more principal than remaining
			if principalPayment > remainingPrincipal {
				principalPayment = remainingPrincipal
			}
		}

		// Create schedule item
		item := models.LoanScheduleItem{
			ScheduledDate:   paymentDate,
			PrincipalAmount: math.Round(principalPayment*100) / 100, // Round to 2 decimal places
			InterestAmount:  math.Round(interestForPeriod*100) / 100,
			TotalAmount:     math.Round((principalPayment+interestForPeriod)*100) / 100,
			Status:          "Pending",
		}

		scheduleItems = append(scheduleItems, item)

		// Update for next iteration
		remainingPrincipal -= principalPayment
		currentDate = paymentDate
	}

	return scheduleItems, nil
}

// calculateTermInMonths calculates the number of months between two dates
func (s *Service) calculateTermInMonths(startDate, endDate time.Time) int {
	years := endDate.Year() - startDate.Year()
	months := int(endDate.Month() - startDate.Month())

	return years*12 + months
}

// addMonths adds the specified number of months to a date
func (s *Service) addMonths(date time.Time, months int) time.Time {
	year := date.Year()
	month := date.Month()

	// Add months
	month = month + time.Month(months)

	// Adjust year if needed
	for month > 12 {
		month -= 12
		year++
	}

	// Create new date with same day (or end of month if day doesn't exist)
	newDate := time.Date(year, month, date.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location())

	// If the day is different (e.g., trying to use Feb 30), it rolled over to the next month
	// So go back to the last day of the intended month
	if newDate.Month() != month {
		// Go to first day of the month that was created
		newDate = time.Date(newDate.Year(), newDate.Month(), 1, date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location())
		// Then go back one day to get the last day of the intended month
		newDate = newDate.AddDate(0, 0, -1)
	}

	return newDate
}

// setDayOfMonth sets the day of the month for a date, handling month boundaries
func (s *Service) setDayOfMonth(date time.Time, day int) time.Time {
	// Get the last day of the month
	lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location()).Day()

	// If requested day is beyond the last day, use the last day
	if day > lastDay {
		day = lastDay
	}

	return time.Date(date.Year(), date.Month(), day, date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location())
}

// GetLoan retrieves a loan by ID
func (s *Service) GetLoan(ctx context.Context, id int64) (*models.Loan, error) {
	query := `
		SELECT id, loan_number, lender_company_id, borrower_company_id,
			   principal_amount, outstanding_amount, interest_rate,
			   start_date, maturity_date, payment_frequency, payment_day,
			   status, collateral_details, notes, created_at, updated_at
		FROM loans
		WHERE id = ?
	`
	loan := &models.Loan{}
	err := s.db.DB().QueryRowContext(ctx, query, id).Scan(
		&loan.ID,
		&loan.LoanNumber,
		&loan.LenderCompanyID,
		&loan.BorrowerCompanyID,
		&loan.PrincipalAmount,
		&loan.OutstandingAmount,
		&loan.InterestRate,
		&loan.StartDate,
		&loan.MaturityDate,
		&loan.PaymentFrequency,
		&loan.PaymentDay,
		&loan.Status,
		&loan.CollateralDetails,
		&loan.Notes,
		&loan.CreatedAt,
		&loan.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("loan not found")
		}
		return nil, fmt.Errorf("error querying loan: %v", err)
	}
	return loan, nil
}

// GetLoanSchedule retrieves the payment schedule for a loan
func (s *Service) GetLoanSchedule(ctx context.Context, loanID int64) ([]models.LoanScheduleItem, error) {
	query := `
		SELECT id, loan_id, scheduled_date, principal_amount, interest_amount,
			   total_amount, status, actual_payment_id, created_at, updated_at
		FROM loan_schedule_items
		WHERE loan_id = ?
		ORDER BY scheduled_date
	`
	rows, err := s.db.DB().QueryContext(ctx, query, loanID)
	if err != nil {
		return nil, fmt.Errorf("error querying loan schedule: %v", err)
	}
	defer rows.Close()

	var scheduleItems []models.LoanScheduleItem
	for rows.Next() {
		var item models.LoanScheduleItem
		var actualPaymentID sql.NullInt64

		err := rows.Scan(
			&item.ID,
			&item.LoanID,
			&item.ScheduledDate,
			&item.PrincipalAmount,
			&item.InterestAmount,
			&item.TotalAmount,
			&item.Status,
			&actualPaymentID,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning loan schedule item: %v", err)
		}

		if actualPaymentID.Valid {
			paymentID := actualPaymentID.Int64
			item.ActualPaymentID = &paymentID
		}

		scheduleItems = append(scheduleItems, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating loan schedule items: %v", err)
	}

	return scheduleItems, nil
}

// RecordLoanPayment records a payment against a loan and updates the loan's outstanding balance
func (s *Service) RecordLoanPayment(ctx context.Context, payment *models.LoanPayment, scheduleItemID int64) error {
	// Start transaction
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Get the loan to verify it exists and check its outstanding amount
	loan, err := s.getLoanForUpdate(ctx, tx, payment.LoanID)
	if err != nil {
		return err
	}

	// Validate payment amount
	if payment.TotalAmount <= 0 {
		return errors.New("payment amount must be greater than zero")
	}

	if payment.TotalAmount != (payment.PrincipalAmount + payment.InterestAmount) {
		return errors.New("total payment amount must equal principal amount plus interest amount")
	}

	// Ensure we're not paying more than outstanding
	if payment.PrincipalAmount > loan.OutstandingAmount {
		return fmt.Errorf("principal payment amount (%f) exceeds outstanding loan amount (%f)", payment.PrincipalAmount, loan.OutstandingAmount)
	}

	// Get the schedule item to update
	scheduleItem, err := s.getScheduleItemForUpdate(ctx, tx, scheduleItemID)
	if err != nil {
		return err
	}

	// Verify schedule item belongs to this loan
	if scheduleItem.LoanID != payment.LoanID {
		return errors.New("schedule item does not belong to the specified loan")
	}

	// Check if schedule item is already paid
	if scheduleItem.Status == "Paid" {
		return errors.New("this schedule item has already been paid")
	}

	// Insert payment record
	query := `
		INSERT INTO loan_payments (
			loan_id, payment_number, payment_date, principal_amount,
			interest_amount, total_amount, payment_method,
			transaction_reference, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		payment.LoanID,
		payment.PaymentNumber,
		payment.PaymentDate,
		payment.PrincipalAmount,
		payment.InterestAmount,
		payment.TotalAmount,
		payment.PaymentMethod,
		payment.TransactionReference,
		payment.Notes,
	)
	if err != nil {
		return fmt.Errorf("error creating payment record: %v", err)
	}

	paymentID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting payment ID: %v", err)
	}
	payment.ID = paymentID

	// Update schedule item to mark it as paid
	query = `
		UPDATE loan_schedule_items
		SET status = ?, actual_payment_id = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err = tx.ExecContext(ctx, query, "Paid", paymentID, scheduleItemID)
	if err != nil {
		return fmt.Errorf("error updating schedule item: %v", err)
	}

	// Update loan outstanding amount
	newOutstandingAmount := loan.OutstandingAmount - payment.PrincipalAmount
	query = `
		UPDATE loans
		SET outstanding_amount = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err = tx.ExecContext(ctx, query, newOutstandingAmount, loan.ID)
	if err != nil {
		return fmt.Errorf("error updating loan outstanding amount: %v", err)
	}

	// If loan is fully paid, update status
	if newOutstandingAmount <= 0 {
		query = `
			UPDATE loans
			SET status = 'Completed', updated_at = NOW()
			WHERE id = ?
		`
		_, err = tx.ExecContext(ctx, query, loan.ID)
		if err != nil {
			return fmt.Errorf("error updating loan status: %v", err)
		}
	}

	return tx.Commit()
}

// getLoanForUpdate gets a loan with row locking for update
func (s *Service) getLoanForUpdate(ctx context.Context, tx *sql.Tx, loanID int64) (*models.Loan, error) {
	query := `
		SELECT id, loan_number, lender_company_id, borrower_company_id,
			   principal_amount, outstanding_amount, interest_rate,
			   start_date, maturity_date, payment_frequency, payment_day,
			   status, collateral_details, notes, created_at, updated_at
		FROM loans
		WHERE id = ?
		FOR UPDATE
	`
	loan := &models.Loan{}
	err := tx.QueryRowContext(ctx, query, loanID).Scan(
		&loan.ID,
		&loan.LoanNumber,
		&loan.LenderCompanyID,
		&loan.BorrowerCompanyID,
		&loan.PrincipalAmount,
		&loan.OutstandingAmount,
		&loan.InterestRate,
		&loan.StartDate,
		&loan.MaturityDate,
		&loan.PaymentFrequency,
		&loan.PaymentDay,
		&loan.Status,
		&loan.CollateralDetails,
		&loan.Notes,
		&loan.CreatedAt,
		&loan.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("loan not found")
		}
		return nil, fmt.Errorf("error querying loan: %v", err)
	}
	return loan, nil
}

// getScheduleItemForUpdate gets a schedule item with row locking for update
func (s *Service) getScheduleItemForUpdate(ctx context.Context, tx *sql.Tx, itemID int64) (*models.LoanScheduleItem, error) {
	query := `
		SELECT id, loan_id, scheduled_date, principal_amount, interest_amount,
			   total_amount, status, actual_payment_id, created_at, updated_at
		FROM loan_schedule_items
		WHERE id = ?
		FOR UPDATE
	`
	var item models.LoanScheduleItem
	var actualPaymentID sql.NullInt64

	err := tx.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID,
		&item.LoanID,
		&item.ScheduledDate,
		&item.PrincipalAmount,
		&item.InterestAmount,
		&item.TotalAmount,
		&item.Status,
		&actualPaymentID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("schedule item not found")
		}
		return nil, fmt.Errorf("error querying schedule item: %v", err)
	}

	if actualPaymentID.Valid {
		paymentID := actualPaymentID.Int64
		item.ActualPaymentID = &paymentID
	}

	return &item, nil
}

// ListLoansByCompany retrieves all loans where a company is either the lender or borrower
func (s *Service) ListLoansByCompany(ctx context.Context, companyID int64, role string) ([]*models.Loan, error) {
	var query string

	switch role {
	case "lender":
		query = `
			SELECT id, loan_number, lender_company_id, borrower_company_id,
				   principal_amount, outstanding_amount, interest_rate,
				   start_date, maturity_date, payment_frequency, payment_day,
				   status, collateral_details, notes, created_at, updated_at
			FROM loans
			WHERE lender_company_id = ?
			ORDER BY created_at DESC
		`
	case "borrower":
		query = `
			SELECT id, loan_number, lender_company_id, borrower_company_id,
				   principal_amount, outstanding_amount, interest_rate,
				   start_date, maturity_date, payment_frequency, payment_day,
				   status, collateral_details, notes, created_at, updated_at
			FROM loans
			WHERE borrower_company_id = ?
			ORDER BY created_at DESC
		`
	case "any":
		query = `
			SELECT id, loan_number, lender_company_id, borrower_company_id,
				   principal_amount, outstanding_amount, interest_rate,
				   start_date, maturity_date, payment_frequency, payment_day,
				   status, collateral_details, notes, created_at, updated_at
			FROM loans
			WHERE lender_company_id = ? OR borrower_company_id = ?
			ORDER BY created_at DESC
		`
	default:
		return nil, fmt.Errorf("invalid role specified: %s", role)
	}

	var rows *sql.Rows
	var err error

	if role == "any" {
		rows, err = s.db.DB().QueryContext(ctx, query, companyID, companyID)
	} else {
		rows, err = s.db.DB().QueryContext(ctx, query, companyID)
	}

	if err != nil {
		return nil, fmt.Errorf("error querying loans: %v", err)
	}
	defer rows.Close()

	var loans []*models.Loan
	for rows.Next() {
		loan := &models.Loan{}
		err := rows.Scan(
			&loan.ID,
			&loan.LoanNumber,
			&loan.LenderCompanyID,
			&loan.BorrowerCompanyID,
			&loan.PrincipalAmount,
			&loan.OutstandingAmount,
			&loan.InterestRate,
			&loan.StartDate,
			&loan.MaturityDate,
			&loan.PaymentFrequency,
			&loan.PaymentDay,
			&loan.Status,
			&loan.CollateralDetails,
			&loan.Notes,
			&loan.CreatedAt,
			&loan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning loan: %v", err)
		}
		loans = append(loans, loan)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating loans: %v", err)
	}

	return loans, nil
}

// UpdateLoanStatus updates the status of a loan
func (s *Service) UpdateLoanStatus(ctx context.Context, loanID int64, newStatus string) error {
	// Validate status
	validStatuses := map[string]bool{
		"Pending":   true,
		"Active":    true,
		"Defaulted": true,
		"Completed": true,
	}

	if !validStatuses[newStatus] {
		return fmt.Errorf("invalid loan status: %s", newStatus)
	}

	query := `
		UPDATE loans
		SET status = ?, updated_at = NOW()
		WHERE id = ?
	`
	result, err := s.db.DB().ExecContext(ctx, query, newStatus, loanID)
	if err != nil {
		return fmt.Errorf("error updating loan status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return errors.New("loan not found")
	}

	return nil
}

// UpdateOverdueLoans finds and updates loans with overdue payments
func (s *Service) UpdateOverdueLoans(ctx context.Context) (int, error) {
	// Find all loans with schedule items that are overdue but not marked as overdue
	query := `
		UPDATE loan_schedule_items
		SET status = 'Overdue', updated_at = NOW()
		WHERE status = 'Pending'
		AND scheduled_date < CURRENT_DATE()
	`
	result, err := s.db.DB().ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("error updating overdue loan schedules: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting rows affected: %v", err)
	}

	// Find all active loans with at least one overdue payment and mark them as defaulted
	// if the payment is more than 30 days late
	query = `
		UPDATE loans l
		SET status = 'Defaulted', updated_at = NOW()
		WHERE status = 'Active'
		AND EXISTS (
			SELECT 1 FROM loan_schedule_items si
			WHERE si.loan_id = l.id
			AND si.status = 'Overdue'
			AND si.scheduled_date < DATE_SUB(CURRENT_DATE(), INTERVAL 30 DAY)
		)
	`
	_, err = s.db.DB().ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("error updating defaulted loans: %v", err)
	}

	return int(rowsAffected), nil
}
