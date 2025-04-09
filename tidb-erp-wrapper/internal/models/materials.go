package models

import "time"

type Warehouse struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StorageLocation struct {
	ID          int64     `json:"id"`
	WarehouseID int64     `json:"warehouse_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	StorageType string    `json:"storage_type"`
	Capacity    float64   `json:"capacity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Product struct {
	ID            int64     `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	UnitOfMeasure string    `json:"unit_of_measure"`
	Weight        float64   `json:"weight"`
	Volume        float64   `json:"volume"`
	MinStockLevel float64   `json:"min_stock_level"`
	MaxStockLevel float64   `json:"max_stock_level"`
	ReorderPoint  float64   `json:"reorder_point"`
	LeadTimeDays  int       `json:"lead_time_days"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type InventoryTransaction struct {
	ID                int64     `json:"id"`
	ProductID         int64     `json:"product_id"`
	WarehouseID       int64     `json:"warehouse_id"`
	StorageLocationID int64     `json:"storage_location_id"`
	TransactionType   string    `json:"transaction_type"`
	ReferenceType     string    `json:"reference_type"`
	ReferenceID       int64     `json:"reference_id"`
	Quantity          float64   `json:"quantity"`
	UnitCost          float64   `json:"unit_cost"`
	TransactionDate   time.Time `json:"transaction_date"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type StockCount struct {
	ID          int64     `json:"id"`
	WarehouseID int64     `json:"warehouse_id"`
	CountDate   time.Time `json:"count_date"`
	Status      string    `json:"status"`
	CountedBy   string    `json:"counted_by"`
	VerifiedBy  string    `json:"verified_by"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StockCountItem struct {
	ID                int64     `json:"id"`
	StockCountID      int64     `json:"stock_count_id"`
	ProductID         int64     `json:"product_id"`
	StorageLocationID int64     `json:"storage_location_id"`
	SystemQuantity    float64   `json:"system_quantity"`
	CountedQuantity   float64   `json:"counted_quantity"`
	VarianceQuantity  float64   `json:"variance_quantity"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
