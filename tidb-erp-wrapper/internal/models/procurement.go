package models

import "time"

type Supplier struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	TaxID       string    `json:"tax_id"`
	Address     string    `json:"address"`
	ContactInfo string    `json:"contact_info"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PurchaseOrder struct {
	ID           int64     `json:"id"`
	OrderNumber  string    `json:"order_number"`
	SupplierID   int64     `json:"supplier_id"`
	OrderDate    time.Time `json:"order_date"`
	DeliveryDate time.Time `json:"delivery_date"`
	Status       string    `json:"status"`
	TotalAmount  float64   `json:"total_amount"`
	CurrencyCode string    `json:"currency_code"`
	PaymentTerms string    `json:"payment_terms"`
	ApprovedBy   string    `json:"approved_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PurchaseOrderItem struct {
	ID              int64     `json:"id"`
	PurchaseOrderID int64     `json:"purchase_order_id"`
	ProductCode     string    `json:"product_code"`
	Description     string    `json:"description"`
	Quantity        float64   `json:"quantity"`
	UnitPrice       float64   `json:"unit_price"`
	TotalPrice      float64   `json:"total_price"`
	TaxRate         float64   `json:"tax_rate"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
