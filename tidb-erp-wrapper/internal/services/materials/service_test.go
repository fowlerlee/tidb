package materials

import (
	"context"
	"testing"
	"time"

	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaterialsService(t *testing.T) {
	testDB, err := testutil.NewTestDB()
	require.NoError(t, err)
	defer testDB.Cleanup()

	svc := NewService(testDB.DB)
	ctx := context.Background()

	t.Run("CreateProduct", func(t *testing.T) {
		product := &models.Product{
			Code:          "PROD001",
			Name:          "Test Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 10.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  20.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}

		err = svc.CreateProduct(ctx, product)
		assert.NoError(t, err)
		assert.NotZero(t, product.ID)

		// Test creating duplicate product code
		dupProduct := &models.Product{
			Code:          "PROD001",
			Name:          "Duplicate Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 10.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  20.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, dupProduct)
		assert.Error(t, err)
	})

	t.Run("CreateWarehouse", func(t *testing.T) {
		warehouse := &models.Warehouse{
			Code:    "WH001",
			Name:    "Test Warehouse",
			Address: "123 Test St",
			Status:  "active",
		}

		err = svc.CreateWarehouse(ctx, warehouse)
		assert.NoError(t, err)
		assert.NotZero(t, warehouse.ID)

		// Test creating duplicate warehouse code
		dupWarehouse := &models.Warehouse{
			Code:    "WH001",
			Name:    "Duplicate Warehouse",
			Address: "456 Test St",
			Status:  "active",
		}
		err = svc.CreateWarehouse(ctx, dupWarehouse)
		assert.Error(t, err)
	})

	t.Run("CreateInventoryTransaction", func(t *testing.T) {
		// Create test warehouse and storage location
		warehouse := &models.Warehouse{
			Code:    "WH002",
			Name:    "Transaction Test Warehouse",
			Address: "123 Test St",
			Status:  "active",
		}
		err := svc.CreateWarehouse(ctx, warehouse)
		require.NoError(t, err)

		location := &models.StorageLocation{
			WarehouseID: warehouse.ID,
			Code:        "LOC001",
			Name:        "Test Location",
			StorageType: "shelf",
			Capacity:    1000.0,
			Status:      "active",
		}
		err = svc.CreateStorageLocation(ctx, location)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "PROD003",
			Name:          "Transaction Test Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 10.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  20.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Test creating a valid inventory transaction
		transaction := &models.InventoryTransaction{
			ProductID:         product.ID,
			WarehouseID:       warehouse.ID,
			StorageLocationID: location.ID,
			TransactionType:   "receipt",
			ReferenceType:     "purchase_order",
			ReferenceID:       1,
			Quantity:          100.0,
			UnitCost:          90.0,
			TransactionDate:   time.Now(),
		}

		err = svc.CreateInventoryTransaction(ctx, transaction)
		assert.NoError(t, err)
		assert.NotZero(t, transaction.ID)

		// Test issuing more than available stock
		invalidTransaction := &models.InventoryTransaction{
			ProductID:         product.ID,
			WarehouseID:       warehouse.ID,
			StorageLocationID: location.ID,
			TransactionType:   "issue",
			ReferenceType:     "sales_order",
			ReferenceID:       1,
			Quantity:          -200.0, // Negative quantity for issue
			UnitCost:          90.0,
			TransactionDate:   time.Now(),
		}

		err = svc.CreateInventoryTransaction(ctx, invalidTransaction)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient stock")
	})

	t.Run("StockCount", func(t *testing.T) {
		// Create test warehouse
		warehouse := &models.Warehouse{
			Code:    "WH003",
			Name:    "Stock Count Warehouse",
			Address: "123 Test St",
			Status:  "active",
		}
		err := svc.CreateWarehouse(ctx, warehouse)
		require.NoError(t, err)

		// Create storage location
		location := &models.StorageLocation{
			WarehouseID: warehouse.ID,
			Code:        "LOC002",
			Name:        "Stock Count Location",
			StorageType: "shelf",
			Capacity:    1000.0,
			Status:      "active",
		}
		err = svc.CreateStorageLocation(ctx, location)
		require.NoError(t, err)

		// Create test product
		product := &models.Product{
			Code:          "PROD004",
			Name:          "Stock Count Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 10.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  20.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Add initial stock
		initialStock := &models.InventoryTransaction{
			ProductID:         product.ID,
			WarehouseID:       warehouse.ID,
			StorageLocationID: location.ID,
			TransactionType:   "receipt",
			ReferenceType:     "purchase_order",
			ReferenceID:       1,
			Quantity:          50.0,
			UnitCost:          10.0,
			TransactionDate:   time.Now(),
		}
		err = svc.CreateInventoryTransaction(ctx, initialStock)
		require.NoError(t, err)

		// Create stock count
		stockCount := &models.StockCount{
			WarehouseID: warehouse.ID,
			CountDate:   time.Now(),
			Status:      "pending",
			CountedBy:   "Tester",
			Notes:       "Test stock count",
		}

		countItems := []models.StockCountItem{
			{
				ProductID:         product.ID,
				StorageLocationID: location.ID,
				CountedQuantity:   45.0, // Shortage of 5 units
				Notes:             "Found 5 units damaged",
			},
		}

		err = svc.CreateStockCount(ctx, stockCount, countItems)
		assert.NoError(t, err)
		assert.NotZero(t, stockCount.ID)
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		// Create test warehouse and location
		warehouse := &models.Warehouse{
			Code:    "WH004",
			Name:    "Rollback Test Warehouse",
			Address: "123 Test St",
			Status:  "active",
		}
		err := svc.CreateWarehouse(ctx, warehouse)
		require.NoError(t, err)

		location := &models.StorageLocation{
			WarehouseID: warehouse.ID,
			Code:        "LOC003",
			Name:        "Rollback Test Location",
			StorageType: "shelf",
			Capacity:    1000.0,
			Status:      "active",
		}
		err = svc.CreateStorageLocation(ctx, location)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "PROD005",
			Name:          "Rollback Test Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 10.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  20.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Try to create an invalid transaction to trigger rollback
		invalidTransaction := &models.InventoryTransaction{
			ProductID:         product.ID,
			WarehouseID:       warehouse.ID,
			StorageLocationID: location.ID,
			TransactionType:   "issue",
			ReferenceType:     "sales_order",
			ReferenceID:       1,
			Quantity:          -100.0, // Negative quantity for issue with no stock
			UnitCost:          90.0,
			TransactionDate:   time.Now(),
		}

		err = svc.CreateInventoryTransaction(ctx, invalidTransaction)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient stock")

		// Verify no inventory transactions were created
		var count int
		err = testDB.DB.DB().QueryRowContext(ctx,
			"SELECT COUNT(*) FROM inventory_transactions WHERE product_id = ? AND warehouse_id = ?",
			product.ID, warehouse.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("GetProductStock", func(t *testing.T) {
		// Create test warehouse
		warehouse := &models.Warehouse{
			Code:    "WH-STOCK",
			Name:    "Stock Test Warehouse",
			Address: "123 Stock St",
			Status:  "active",
		}
		err = svc.CreateWarehouse(ctx, warehouse)
		require.NoError(t, err)

		location := &models.StorageLocation{
			WarehouseID: warehouse.ID,
			Code:        "LOC-STOCK",
			Name:        "Stock Test Location",
			StorageType: "shelf",
			Capacity:    1000.0,
			Status:      "active",
		}
		err = svc.CreateStorageLocation(ctx, location)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "STOCK-TEST",
			Name:          "Stock Test Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 10.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  20.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Create an inbound transaction
		transaction := &models.InventoryTransaction{
			ProductID:         product.ID,
			WarehouseID:       warehouse.ID,
			StorageLocationID: location.ID,
			TransactionType:   "receipt",
			ReferenceType:     "purchase_order",
			ReferenceID:       1,
			Quantity:          50.0,
			UnitCost:          10.0,
			TransactionDate:   time.Now(),
		}
		err = svc.CreateInventoryTransaction(ctx, transaction)
		require.NoError(t, err)

		// Create an outbound transaction
		outbound := &models.InventoryTransaction{
			ProductID:         product.ID,
			WarehouseID:       warehouse.ID,
			StorageLocationID: location.ID,
			TransactionType:   "issue",
			ReferenceType:     "sales_order",
			ReferenceID:       2,
			Quantity:          -20.0,
			UnitCost:          10.0,
			TransactionDate:   time.Now(),
		}
		err = svc.CreateInventoryTransaction(ctx, outbound)
		require.NoError(t, err)

		// Get stock levels
		stock, err := svc.GetProductStock(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, 30.0, stock[warehouse.ID])
	})

	t.Run("GetLowStockProducts", func(t *testing.T) {
		// Create test warehouse
		warehouse := &models.Warehouse{
			Code:    "WH-LOW",
			Name:    "Low Stock Warehouse",
			Address: "123 Low St",
			Status:  "active",
		}
		err = svc.CreateWarehouse(ctx, warehouse)
		require.NoError(t, err)

		location := &models.StorageLocation{
			WarehouseID: warehouse.ID,
			Code:        "LOC-LOW",
			Name:        "Low Stock Location",
			StorageType: "shelf",
			Capacity:    1000.0,
			Status:      "active",
		}
		err = svc.CreateStorageLocation(ctx, location)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "PROD-LOW",
			Name:          "Low Stock Product",
			Description:   "Test Description",
			Category:      "Test Category",
			UnitOfMeasure: "PCS",
			Weight:        1.5,
			Volume:        2.0,
			MinStockLevel: 50.0,
			MaxStockLevel: 100.0,
			ReorderPoint:  75.0,
			LeadTimeDays:  7,
			IsActive:      true,
		}
		err = svc.CreateProduct(ctx, product)
		require.NoError(t, err)

		// Add stock below reorder point
		transaction := &models.InventoryTransaction{
			ProductID:         product.ID,
			WarehouseID:       warehouse.ID,
			StorageLocationID: location.ID,
			TransactionType:   "receipt",
			ReferenceType:     "purchase_order",
			ReferenceID:       1,
			Quantity:          30.0,
			UnitCost:          90.0,
			TransactionDate:   time.Now(),
		}
		err = svc.CreateInventoryTransaction(ctx, transaction)
		require.NoError(t, err)

		// Check low stock products
		lowStockProducts, err := svc.GetLowStockProducts(ctx)
		require.NoError(t, err)
		found := false
		for _, p := range lowStockProducts {
			if p.ID == product.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Low stock product should be in the results")
	})
}
