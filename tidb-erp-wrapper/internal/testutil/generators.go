package testutil

import (
	"fmt"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
	"time"
)

// GenerateSupplier creates a test supplier
func GenerateSupplier(id int64) *models.Supplier {
	return &models.Supplier{
		ID:          id,
		Name:        fmt.Sprintf("Test Supplier %d", id),
		Code:        fmt.Sprintf("SUP%d", id),
		TaxID:       fmt.Sprintf("TAX%d", id),
		Address:     "123 Test St",
		ContactInfo: "contact@test.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// GenerateProduct creates a test product
func GenerateProduct(id int64) *models.Product {
	return &models.Product{
		ID:            id,
		Code:          fmt.Sprintf("PRD%d", id),
		Name:          fmt.Sprintf("Test Product %d", id),
		Description:   "Test product description",
		Category:      "Test Category",
		UnitOfMeasure: "PCS",
		Weight:        1.0,
		Volume:        1.0,
		MinStockLevel: 10,
		MaxStockLevel: 100,
		ReorderPoint:  20,
		LeadTimeDays:  7,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// GenerateEmployee creates a test employee
func GenerateEmployee(id int64) *models.Employee {
	return &models.Employee{
		ID:             id,
		EmployeeNumber: fmt.Sprintf("EMP%d", id),
		FirstName:      fmt.Sprintf("First%d", id),
		LastName:       fmt.Sprintf("Last%d", id),
		Email:          fmt.Sprintf("emp%d@test.com", id),
		Phone:          "1234567890",
		HireDate:       time.Now(),
		DepartmentID:   1,
		JobPositionID:  1,
		Status:         "active",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// GenerateCustomer creates a test customer
func GenerateCustomer(id int64) *models.Customer {
	return &models.Customer{
		ID:           id,
		Name:         fmt.Sprintf("Test Customer %d", id),
		Code:         fmt.Sprintf("CUS%d", id),
		TaxID:        fmt.Sprintf("TAX%d", id),
		Address:      "456 Test Ave",
		ContactInfo:  "customer@test.com",
		PaymentTerms: "Net 30",
		CreditLimit:  10000.0,
		CurrencyCode: "USD",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// GeneratePurchaseOrder creates a test purchase order
func GeneratePurchaseOrder(id int64, supplierID int64) *models.PurchaseOrder {
	return &models.PurchaseOrder{
		ID:           id,
		OrderNumber:  fmt.Sprintf("PO%d", id),
		SupplierID:   supplierID,
		OrderDate:    time.Now(),
		DeliveryDate: time.Now().Add(7 * 24 * time.Hour),
		Status:       "draft",
		TotalAmount:  1000.0,
		CurrencyCode: "USD",
		PaymentTerms: "Net 30",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// GeneratePurchaseOrderItem creates a test purchase order item
func GeneratePurchaseOrderItem(id int64, poID int64) *models.PurchaseOrderItem {
	return &models.PurchaseOrderItem{
		ID:              id,
		PurchaseOrderID: poID,
		ProductCode:     fmt.Sprintf("PRD%d", id),
		Description:     "Test product",
		Quantity:        10,
		UnitPrice:       100.0,
		TotalPrice:      1000.0,
		TaxRate:         0.1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// GenerateQualityParameter creates a test quality parameter
func GenerateQualityParameter(id int64) *models.QualityParameter {
	minValue := 0.0
	maxValue := 100.0
	targetValue := 50.0
	return &models.QualityParameter{
		ID:              id,
		Code:            fmt.Sprintf("QP%d", id),
		Name:            fmt.Sprintf("Test Parameter %d", id),
		Description:     "Test parameter description",
		MeasurementUnit: "mm",
		MinValue:        &minValue,
		MaxValue:        &maxValue,
		TargetValue:     &targetValue,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// GenerateInspectionPlan creates a test inspection plan
func GenerateInspectionPlan(id int64) *models.InspectionPlan {
	return &models.InspectionPlan{
		ID:            id,
		Code:          fmt.Sprintf("IP%d", id),
		Name:          fmt.Sprintf("Test Plan %d", id),
		Description:   "Test plan description",
		Version:       "1.0",
		Status:        "active",
		EffectiveFrom: time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// GenerateDepartment creates a test department
func GenerateDepartment(id int64) *models.Department {
	return &models.Department{
		ID:         id,
		Code:       fmt.Sprintf("DEPT%d", id),
		Name:       fmt.Sprintf("Test Department %d", id),
		CostCenter: fmt.Sprintf("CC%d", id),
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// GenerateAccount creates a test account
func GenerateAccount(id int64) *models.Account {
	return &models.Account{
		ID:           id,
		AccountCode:  fmt.Sprintf("ACC%d", id),
		Name:         fmt.Sprintf("Test Account %d", id),
		Type:         "Asset",
		SubType:      "Current Asset",
		Description:  "Test account description",
		Balance:      1000.0,
		CurrencyCode: "USD",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// GenerateJournalEntry creates a test journal entry
func GenerateJournalEntry(id int64) *models.JournalEntry {
	return &models.JournalEntry{
		ID:          id,
		EntryNumber: fmt.Sprintf("JE%d", id),
		Date:        time.Now(),
		Reference:   fmt.Sprintf("REF%d", id),
		Description: "Test journal entry",
		Status:      "draft",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
