package models

import "time"

type Customer struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	TaxID        string    `json:"tax_id"`
	Address      string    `json:"address"`
	ContactInfo  string    `json:"contact_info"`
	PaymentTerms string    `json:"payment_terms"`
	CreditLimit  float64   `json:"credit_limit"`
	CurrencyCode string    `json:"currency_code"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PriceList struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	CurrencyCode string    `json:"currency_code"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PriceListItem struct {
	ID              int64     `json:"id"`
	PriceListID     int64     `json:"price_list_id"`
	ProductCode     string    `json:"product_code"`
	UnitPrice       float64   `json:"unit_price"`
	MinQuantity     float64   `json:"min_quantity"`
	DiscountPercent float64   `json:"discount_percentage"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SalesOrder struct {
	ID                    int64     `json:"id"`
	OrderNumber           string    `json:"order_number"`
	CustomerID            int64     `json:"customer_id"`
	PriceListID           int64     `json:"price_list_id"`
	OrderDate             time.Time `json:"order_date"`
	RequestedDeliveryDate time.Time `json:"requested_delivery_date"`
	Status                string    `json:"status"`
	TotalAmount           float64   `json:"total_amount"`
	TaxAmount             float64   `json:"tax_amount"`
	CurrencyCode          string    `json:"currency_code"`
	PaymentTerms          string    `json:"payment_terms"`
	SalesPerson           string    `json:"sales_person"`
	Notes                 string    `json:"notes"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type SalesOrderItem struct {
	ID              int64     `json:"id"`
	SalesOrderID    int64     `json:"sales_order_id"`
	ProductCode     string    `json:"product_code"`
	Description     string    `json:"description"`
	Quantity        float64   `json:"quantity"`
	UnitPrice       float64   `json:"unit_price"`
	DiscountPercent float64   `json:"discount_percentage"`
	TotalPrice      float64   `json:"total_price"`
	TaxRate         float64   `json:"tax_rate"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Delivery struct {
	ID              int64     `json:"id"`
	DeliveryNumber  string    `json:"delivery_number"`
	SalesOrderID    int64     `json:"sales_order_id"`
	DeliveryDate    time.Time `json:"delivery_date"`
	Status          string    `json:"status"`
	ShippingAddress string    `json:"shipping_address"`
	TrackingNumber  string    `json:"tracking_number"`
	Carrier         string    `json:"carrier"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type DeliveryItem struct {
	ID               int64     `json:"id"`
	DeliveryID       int64     `json:"delivery_id"`
	SalesOrderItemID int64     `json:"sales_order_item_id"`
	Quantity         float64   `json:"quantity"`
	BatchNumber      string    `json:"batch_number"`
	SerialNumbers    string    `json:"serial_numbers"`
	Notes            string    `json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
