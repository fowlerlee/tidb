package materials

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

func (s *Service) CreateWarehouse(ctx context.Context, warehouse *models.Warehouse) error {
	query := `
		INSERT INTO warehouses (
			code, name, address, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		warehouse.Code,
		warehouse.Name,
		warehouse.Address,
		warehouse.Status,
	)
	if err != nil {
		return fmt.Errorf("error creating warehouse: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	warehouse.ID = id
	return nil
}

func (s *Service) CreateStorageLocation(ctx context.Context, location *models.StorageLocation) error {
	query := `
		INSERT INTO storage_locations (
			warehouse_id, code, name, storage_type,
			capacity, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		location.WarehouseID,
		location.Code,
		location.Name,
		location.StorageType,
		location.Capacity,
		location.Status,
	)
	if err != nil {
		return fmt.Errorf("error creating storage location: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	location.ID = id
	return nil
}

func (s *Service) CreateProduct(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (
			code, name, description, category,
			unit_of_measure, weight, volume,
			min_stock_level, max_stock_level,
			reorder_point, lead_time_days,
			is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		product.Code,
		product.Name,
		product.Description,
		product.Category,
		product.UnitOfMeasure,
		product.Weight,
		product.Volume,
		product.MinStockLevel,
		product.MaxStockLevel,
		product.ReorderPoint,
		product.LeadTimeDays,
		product.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating product: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	product.ID = id
	return nil
}

func (s *Service) CreateInventoryTransaction(ctx context.Context, transaction *models.InventoryTransaction) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Check current stock level
	var currentStock float64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(quantity), 0)
		FROM inventory_transactions
		WHERE product_id = ?
		AND warehouse_id = ?
		AND storage_location_id = ?
	`, transaction.ProductID, transaction.WarehouseID, transaction.StorageLocationID).Scan(&currentStock)
	if err != nil {
		return fmt.Errorf("error getting current stock: %v", err)
	}

	// For outbound transactions, check if enough stock is available
	if transaction.Quantity < 0 && (currentStock+transaction.Quantity) < 0 {
		return errors.New("insufficient stock for outbound transaction")
	}

	// Insert inventory transaction
	query := `
		INSERT INTO inventory_transactions (
			product_id, warehouse_id, storage_location_id,
			transaction_type, reference_type, reference_id,
			quantity, unit_cost, transaction_date,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		transaction.ProductID,
		transaction.WarehouseID,
		transaction.StorageLocationID,
		transaction.TransactionType,
		transaction.ReferenceType,
		transaction.ReferenceID,
		transaction.Quantity,
		transaction.UnitCost,
		transaction.TransactionDate,
	)
	if err != nil {
		return fmt.Errorf("error creating inventory transaction: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	transaction.ID = id

	return tx.Commit()
}

func (s *Service) CreateStockCount(ctx context.Context, stockCount *models.StockCount, items []models.StockCountItem) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert stock count
	query := `
		INSERT INTO stock_counts (
			warehouse_id, count_date, status,
			counted_by, verified_by, notes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		stockCount.WarehouseID,
		stockCount.CountDate,
		stockCount.Status,
		stockCount.CountedBy,
		stockCount.VerifiedBy,
		stockCount.Notes,
	)
	if err != nil {
		return fmt.Errorf("error creating stock count: %v", err)
	}

	stockCountID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting stock count ID: %v", err)
	}
	stockCount.ID = stockCountID

	// Insert stock count items
	for _, item := range items {
		// Get system quantity
		var systemQty float64
		err = tx.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(quantity), 0)
			FROM inventory_transactions
			WHERE product_id = ?
			AND warehouse_id = ?
			AND storage_location_id = ?
		`, item.ProductID, stockCount.WarehouseID, item.StorageLocationID).Scan(&systemQty)
		if err != nil {
			return fmt.Errorf("error getting system quantity: %v", err)
		}

		query := `
			INSERT INTO stock_count_items (
				stock_count_id, product_id, storage_location_id,
				system_quantity, counted_quantity, variance_quantity,
				notes, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		`
		_, err := tx.ExecContext(ctx, query,
			stockCountID,
			item.ProductID,
			item.StorageLocationID,
			systemQty,
			item.CountedQuantity,
			item.CountedQuantity-systemQty,
			item.Notes,
		)
		if err != nil {
			return fmt.Errorf("error creating stock count item: %v", err)
		}
	}

	return tx.Commit()
}

func (s *Service) GetProductStock(ctx context.Context, productID int64) (map[int64]float64, error) {
	query := `
		SELECT warehouse_id, COALESCE(SUM(quantity), 0) as stock
		FROM inventory_transactions
		WHERE product_id = ?
		GROUP BY warehouse_id
	`
	rows, err := s.db.DB().QueryContext(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("error querying product stock: %v", err)
	}
	defer rows.Close()

	stock := make(map[int64]float64)
	for rows.Next() {
		var warehouseID int64
		var quantity float64
		if err := rows.Scan(&warehouseID, &quantity); err != nil {
			return nil, fmt.Errorf("error scanning stock result: %v", err)
		}
		stock[warehouseID] = quantity
	}

	return stock, nil
}

func (s *Service) GetLowStockProducts(ctx context.Context) ([]models.Product, error) {
	query := `
		WITH product_stock AS (
			SELECT product_id, COALESCE(SUM(quantity), 0) as current_stock
			FROM inventory_transactions
			GROUP BY product_id
		)
		SELECT p.id, p.code, p.name, p.description,
			   p.category, p.unit_of_measure, p.weight,
			   p.volume, p.min_stock_level, p.max_stock_level,
			   p.reorder_point, p.lead_time_days, p.is_active,
			   p.created_at, p.updated_at
		FROM products p
		JOIN product_stock ps ON p.id = ps.product_id
		WHERE ps.current_stock <= p.reorder_point
		AND p.is_active = true
	`
	rows, err := s.db.DB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying low stock products: %v", err)
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Code,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.UnitOfMeasure,
			&product.Weight,
			&product.Volume,
			&product.MinStockLevel,
			&product.MaxStockLevel,
			&product.ReorderPoint,
			&product.LeadTimeDays,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning product: %v", err)
		}
		products = append(products, product)
	}

	return products, nil
}
