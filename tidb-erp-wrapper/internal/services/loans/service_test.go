package loans

import (
	"context"
	"testing"
	"time"

	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/db"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoanService(t *testing.T) {
	// Setup test DB
	testDB := testutil.SetupTestDB(t)
	defer testDB.Close()

	// Create service instance
	dbHandler := &db.DBHandler{DB: testDB}
	svc := NewService(dbHandler)

	t.Run("CompanyOperations", func(t *testing.T) {
		// Create a company
		company := &models.Company{
			Name:               "Acme Corp",
			RegistrationNumber: "AC123456",
			TaxID:              "TAX123456",
			ContactPerson:      "John Doe",
			Email:              "john@acmecorp.com",
			Phone:              "555-123-4567",
			Address:            "123 Main St, Business City",
			IsActive:           true,
		}

		err := svc.CreateCompany(context.Background(), company)
		require.NoError(t, err)
		require.NotZero(t, company.ID)

		// Get company by ID
		retrievedCompany, err := svc.GetCompany(context.Background(), company.ID)
		assert.NoError(t, err)
		assert.Equal(t, company.Name, retrievedCompany.Name)
		assert.Equal(t, company.RegistrationNumber, retrievedCompany.RegistrationNumber)

		// Get company by registration number
		retrievedByRegNum, err := svc.GetCompanyByRegistrationNumber(context.Background(), company.RegistrationNumber)
		assert.NoError(t, err)
		assert.Equal(t, company.ID, retrievedByRegNum.ID)

		// Non-existent company
		_, err = svc.GetCompany(context.Background(), 9999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("LoanCreationAndRetrieval", func(t *testing.T) {
		// Create two companies for the loan
		lender := &models.Company{
			Name:               "Lender Corp",
			RegistrationNumber: "LC123456",
			TaxID:              "LTAX123456",
			IsActive:           true,
		}
		err := svc.CreateCompany(context.Background(), lender)
		require.NoError(t, err)

		borrower := &models.Company{
			Name:               "Borrower Inc",
			RegistrationNumber: "BI123456",
			TaxID:              "BTAX123456",
			IsActive:           true,
		}
		err = svc.CreateCompany(context.Background(), borrower)
		require.NoError(t, err)

		// Create loan
		now := time.Now()
		loan := &models.Loan{
			LoanNumber:        "L2025-001",
			LenderCompanyID:   lender.ID,
			BorrowerCompanyID: borrower.ID,
			PrincipalAmount:   100000,
			InterestRate:      5.0, // 5% annual interest
			StartDate:         now,
			MaturityDate:      now.AddDate(2, 0, 0), // 2 years
			PaymentFrequency:  "Monthly",
			PaymentDay:        15,
			Status:            "Active",
		}

		err = svc.CreateLoan(context.Background(), loan)
		require.NoError(t, err)
		require.NotZero(t, loan.ID)

		// Retrieve loan
		retrievedLoan, err := svc.GetLoan(context.Background(), loan.ID)
		assert.NoError(t, err)
		assert.Equal(t, loan.LoanNumber, retrievedLoan.LoanNumber)
		assert.Equal(t, loan.PrincipalAmount, retrievedLoan.PrincipalAmount)
		assert.Equal(t, loan.OutstandingAmount, retrievedLoan.OutstandingAmount)

		// Check loan schedule was created
		schedule, err := svc.GetLoanSchedule(context.Background(), loan.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, schedule)

		// Verify expected number of payments (24 months for a 2-year monthly payment loan)
		assert.Len(t, schedule, 24)

		// Verify first payment details
		firstPayment := schedule[0]
		assert.Equal(t, "Pending", firstPayment.Status)
		assert.Equal(t, loan.ID, firstPayment.LoanID)

		// Verify total schedule amounts add up to principal
		var totalPrincipal float64
		for _, item := range schedule {
			totalPrincipal += item.PrincipalAmount
		}
		assert.InDelta(t, loan.PrincipalAmount, totalPrincipal, 0.1) // Account for rounding differences
	})

	t.Run("LoanPaymentProcessing", func(t *testing.T) {
		// Create companies and loan first
		lender := &models.Company{
			Name:               "ABC Lender",
			RegistrationNumber: "ABC123",
			IsActive:           true,
		}
		err := svc.CreateCompany(context.Background(), lender)
		require.NoError(t, err)

		borrower := &models.Company{
			Name:               "XYZ Borrower",
			RegistrationNumber: "XYZ123",
			IsActive:           true,
		}
		err = svc.CreateCompany(context.Background(), borrower)
		require.NoError(t, err)

		now := time.Now()
		loan := &models.Loan{
			LoanNumber:        "LOAN-TEST-1",
			LenderCompanyID:   lender.ID,
			BorrowerCompanyID: borrower.ID,
			PrincipalAmount:   10000,
			InterestRate:      6.0,
			StartDate:         now,
			MaturityDate:      now.AddDate(1, 0, 0), // 1 year
			PaymentFrequency:  "Monthly",
			PaymentDay:        1,
			Status:            "Active",
		}

		err = svc.CreateLoan(context.Background(), loan)
		require.NoError(t, err)

		// Get schedule to process a payment
		schedule, err := svc.GetLoanSchedule(context.Background(), loan.ID)
		require.NoError(t, err)
		require.NotEmpty(t, schedule)

		firstScheduleItem := schedule[0]

		// Create payment for first schedule item
		payment := &models.LoanPayment{
			LoanID:          loan.ID,
			PaymentNumber:   "PAY-001",
			PaymentDate:     now,
			PrincipalAmount: firstScheduleItem.PrincipalAmount,
			InterestAmount:  firstScheduleItem.InterestAmount,
			TotalAmount:     firstScheduleItem.TotalAmount,
			PaymentMethod:   "Bank Transfer",
		}

		// Record payment
		err = svc.RecordLoanPayment(context.Background(), payment, firstScheduleItem.ID)
		assert.NoError(t, err)
		assert.NotZero(t, payment.ID)

		// Check that schedule item is now marked as paid
		updatedSchedule, err := svc.GetLoanSchedule(context.Background(), loan.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Paid", updatedSchedule[0].Status)
		assert.NotNil(t, updatedSchedule[0].ActualPaymentID)
		assert.Equal(t, payment.ID, *updatedSchedule[0].ActualPaymentID)

		// Check loan's outstanding balance was reduced
		updatedLoan, err := svc.GetLoan(context.Background(), loan.ID)
		assert.NoError(t, err)
		assert.InDelta(t, loan.PrincipalAmount-payment.PrincipalAmount, updatedLoan.OutstandingAmount, 0.01)
	})

	t.Run("LoanListingByCompany", func(t *testing.T) {
		// Create companies
		lender := &models.Company{
			Name:               "Mega Lender",
			RegistrationNumber: "ML-001",
			IsActive:           true,
		}
		err := svc.CreateCompany(context.Background(), lender)
		require.NoError(t, err)

		borrower1 := &models.Company{
			Name:               "Small Borrower",
			RegistrationNumber: "SB-001",
			IsActive:           true,
		}
		err = svc.CreateCompany(context.Background(), borrower1)
		require.NoError(t, err)

		borrower2 := &models.Company{
			Name:               "Medium Borrower",
			RegistrationNumber: "MB-001",
			IsActive:           true,
		}
		err = svc.CreateCompany(context.Background(), borrower2)
		require.NoError(t, err)

		// Create two loans with the same lender but different borrowers
		now := time.Now()
		loan1 := &models.Loan{
			LoanNumber:        "LIST-LOAN-1",
			LenderCompanyID:   lender.ID,
			BorrowerCompanyID: borrower1.ID,
			PrincipalAmount:   50000,
			InterestRate:      5.5,
			StartDate:         now,
			MaturityDate:      now.AddDate(3, 0, 0),
			PaymentFrequency:  "Quarterly",
			Status:            "Active",
		}
		err = svc.CreateLoan(context.Background(), loan1)
		require.NoError(t, err)

		loan2 := &models.Loan{
			LoanNumber:        "LIST-LOAN-2",
			LenderCompanyID:   lender.ID,
			BorrowerCompanyID: borrower2.ID,
			PrincipalAmount:   75000,
			InterestRate:      6.0,
			StartDate:         now,
			MaturityDate:      now.AddDate(5, 0, 0),
			PaymentFrequency:  "Quarterly",
			Status:            "Active",
		}
		err = svc.CreateLoan(context.Background(), loan2)
		require.NoError(t, err)

		// List loans by lender
		lenderLoans, err := svc.ListLoansByCompany(context.Background(), lender.ID, "lender")
		assert.NoError(t, err)
		assert.Len(t, lenderLoans, 2)

		// List loans by borrower
		borrower1Loans, err := svc.ListLoansByCompany(context.Background(), borrower1.ID, "borrower")
		assert.NoError(t, err)
		assert.Len(t, borrower1Loans, 1)
		assert.Equal(t, loan1.ID, borrower1Loans[0].ID)

		// List loans by any role
		borrower2Loans, err := svc.ListLoansByCompany(context.Background(), borrower2.ID, "any")
		assert.NoError(t, err)
		assert.Len(t, borrower2Loans, 1)
		assert.Equal(t, loan2.ID, borrower2Loans[0].ID)
	})

	t.Run("LoanStatusUpdate", func(t *testing.T) {
		// Create test loan
		lender := &models.Company{
			Name:               "Status Lender",
			RegistrationNumber: "SL-001",
			IsActive:           true,
		}
		err := svc.CreateCompany(context.Background(), lender)
		require.NoError(t, err)

		borrower := &models.Company{
			Name:               "Status Borrower",
			RegistrationNumber: "SB-001",
			IsActive:           true,
		}
		err = svc.CreateCompany(context.Background(), borrower)
		require.NoError(t, err)

		now := time.Now()
		loan := &models.Loan{
			LoanNumber:        "STATUS-LOAN",
			LenderCompanyID:   lender.ID,
			BorrowerCompanyID: borrower.ID,
			PrincipalAmount:   25000,
			InterestRate:      4.5,
			StartDate:         now,
			MaturityDate:      now.AddDate(2, 0, 0),
			PaymentFrequency:  "Monthly",
			Status:            "Pending",
		}
		err = svc.CreateLoan(context.Background(), loan)
		require.NoError(t, err)

		// Update status
		err = svc.UpdateLoanStatus(context.Background(), loan.ID, "Active")
		assert.NoError(t, err)

		// Verify status update
		updatedLoan, err := svc.GetLoan(context.Background(), loan.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Active", updatedLoan.Status)

		// Try invalid status
		err = svc.UpdateLoanStatus(context.Background(), loan.ID, "Invalid")
		assert.Error(t, err)
	})

	t.Run("EdgeCases", func(t *testing.T) {
		// Create companies
		lenderComp := &models.Company{
			Name:               "Edge Lender",
			RegistrationNumber: "EL-001",
			IsActive:           true,
		}
		err := svc.CreateCompany(context.Background(), lenderComp)
		require.NoError(t, err)

		borrowerComp := &models.Company{
			Name:               "Edge Borrower",
			RegistrationNumber: "EB-001",
			IsActive:           true,
		}
		err = svc.CreateCompany(context.Background(), borrowerComp)
		require.NoError(t, err)

		// Test case 1: Attempt to create loan with invalid data
		invalidLoan := &models.Loan{
			LoanNumber:        "INVALID-LOAN",
			LenderCompanyID:   lenderComp.ID,
			BorrowerCompanyID: lenderComp.ID, // Same company as lender
			PrincipalAmount:   10000,
			InterestRate:      5.0,
			StartDate:         time.Now(),
			MaturityDate:      time.Now().AddDate(1, 0, 0),
		}
		err = svc.CreateLoan(context.Background(), invalidLoan)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lender and borrower cannot be the same company")

		// Test case 2: Negative principal amount
		invalidLoan2 := &models.Loan{
			LoanNumber:        "INVALID-LOAN-2",
			LenderCompanyID:   lenderComp.ID,
			BorrowerCompanyID: borrowerComp.ID,
			PrincipalAmount:   -5000, // Negative amount
			InterestRate:      5.0,
			StartDate:         time.Now(),
			MaturityDate:      time.Now().AddDate(1, 0, 0),
		}
		err = svc.CreateLoan(context.Background(), invalidLoan2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "principal amount must be greater than zero")

		// Test case 3: Start date after maturity date
		now := time.Now()
		invalidLoan3 := &models.Loan{
			LoanNumber:        "INVALID-LOAN-3",
			LenderCompanyID:   lenderComp.ID,
			BorrowerCompanyID: borrowerComp.ID,
			PrincipalAmount:   10000,
			InterestRate:      5.0,
			StartDate:         now.AddDate(2, 0, 0),
			MaturityDate:      now, // Earlier than start date
		}
		err = svc.CreateLoan(context.Background(), invalidLoan3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start date must be before maturity date")
	})
}
