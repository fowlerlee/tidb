package materials

import (
	"context"
	"fmt"
	"testing"
	"tidb-erp-wrapper/internal/models"
	"tidb-erp-wrapper/internal/testutil"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaterialsService(t *testing.T) {
	testDB, err := testutil.NewTestDB()
	require.NoError(t, err)
	defer testDB.Cleanup()

	svc := NewService(testDB.DB)

	t.Run("CreateProductCategory", func(t *testing.T) {
		// Test creating a valid category
		category := &models.ProductCategory{
			Code:        "CAT001",
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    nil,
			IsActive:    true,
		}

		err := svc.CreateProductCategory(context.Background(), category)
		assert.NoError(t, err)
		assert.NotZero(t, category.ID)

		// Test creating duplicate category code
		dupCategory := &models.ProductCategory{
			Code:        "CAT001",
			Name:        "Duplicate Category",
			Description: "Test Description",
			ParentID:    nil,
			IsActive:    true,
		}
		err = svc.CreateProductCategory(context.Background(), dupCategory)
		assert.Error(t, err)

		// Test creating category with invalid parent
		invalidParentCategory := &models.ProductCategory{
			Code:        "CAT002",
			Name:        "Invalid Parent Category",
			Description: "Test Description",
			ParentID:    new(int64),
			IsActive:    true,
		}
		*invalidParentCategory.ParentID = 999999
		err = svc.CreateProductCategory(context.Background(), invalidParentCategory)
		assert.Error(t, err)
	})

	t.Run("CreateProduct", func(t *testing.T) {
		// Create test category first
		category := &models.ProductCategory{
			Code:        "CAT001",
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    nil,
			IsActive:    true,
		}
		err := svc.CreateProductCategory(context.Background(), category)
		require.NoError(t, err)

		// Test creating a valid product
		product := &models.Product{
			Code:          "PROD001",
			Name:          "Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			UnitPrice:     100.0,
			ReorderPoint:  10,
			MinimumOrder:  5,
			LeadTime:      7,
			IsActive:      true,
			Specifications: map[string]interface{}{
				"weight": 1.5,
				"color":  "blue",
			},
		}

		err = svc.CreateProduct(context.Background(), product)
		assert.NoError(t, err)
		assert.NotZero(t, product.ID)

		// Test creating duplicate product code
		dupProduct := &models.Product{
			Code:          "PROD001",
			Name:          "Duplicate Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			UnitPrice:     100.0,
			ReorderPoint:  10,
			MinimumOrder:  5,
			LeadTime:      7,
			IsActive:      true,
		}
		err = svc.CreateProduct(context.Background(), dupProduct)
		assert.Error(t, err)

		// Test creating product with invalid category
		invalidProduct := &models.Product{
			Code:          "PROD002",
			Name:          "Invalid Category Product",
			Description:   "Test Description",
			CategoryID:    999999,
			UnitOfMeasure: "PCS",
			UnitPrice:     100.0,
			ReorderPoint:  10,
			MinimumOrder:  5,
			LeadTime:      7,
			IsActive:      true,
		}
		err = svc.CreateProduct(context.Background(), invalidProduct)
		assert.Error(t, err)
	})

	t.Run("CreateWarehouse", func(t *testing.T) {
		// Test creating a valid warehouse
		warehouse := &models.Warehouse{
			Code:        "WH001",
			Name:        "Test Warehouse",
			Description: "Test Description",
			Address:     "123 Test St",
			IsActive:    true,
			Capacity:    1000.0,
		}

		err := svc.CreateWarehouse(context.Background(), warehouse)
		assert.NoError(t, err)
		assert.NotZero(t, warehouse.ID)

		// Test creating duplicate warehouse code
		dupWarehouse := &models.Warehouse{
			Code:        "WH001",
			Name:        "Duplicate Warehouse",
			Description: "Test Description",
			Address:     "456 Test St",
			IsActive:    true,
			Capacity:    1000.0,
		}
		err = svc.CreateWarehouse(context.Background(), dupWarehouse)
		assert.Error(t, err)
	})

	t.Run("CreateStockMovement", func(t *testing.T) {
		// Create test warehouse
		warehouse := &models.Warehouse{
			Code:        "WH001",
			Name:        "Test Warehouse",
			Description: "Test Description",
			Address:     "123 Test St",
			IsActive:    true,
			Capacity:    1000.0,
		}
		err := svc.CreateWarehouse(context.Background(), warehouse)
		require.NoError(t, err)

		// Create test product
		category := &models.ProductCategory{
			Code:        "CAT001",
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    nil,
			IsActive:    true,
		}
		err = svc.CreateProductCategory(context.Background(), category)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "PROD001",
			Name:          "Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			UnitPrice:     100.0,
			ReorderPoint:  10,
			MinimumOrder:  5,
			LeadTime:      7,
			IsActive:      true,
		}
		err = svc.CreateProduct(context.Background(), product)
		require.NoError(t, err)

		// Test creating a valid stock movement (receipt)
		movement := &models.StockMovement{
			ReferenceNo:  "RCV001",
			MovementType: "receipt",
			WarehouseID:  warehouse.ID,
			MovementDate: time.Now(),
			Status:       "pending",
			Notes:        "Test receipt",
		}

		items := []models.StockMovementItem{
			{
				ProductID:     product.ID,
				Quantity:      100,
				UnitCost:      90.0,
				BatchNumber:   "BATCH001",
				ExpiryDate:    time.Now().AddDate(1, 0, 0),
				SerialNumbers: "SN001-SN100",
			},
		}

		err = svc.CreateStockMovement(context.Background(), movement, items)
		assert.NoError(t, err)
		assert.NotZero(t, movement.ID)

		// Test creating movement with invalid warehouse
		invalidMovement := &models.StockMovement{
			ReferenceNo:  "RCV002",
			MovementType: "receipt",
			WarehouseID:  999999,
			MovementDate: time.Now(),
			Status:       "pending",
			Notes:        "Test invalid receipt",
		}
		err = svc.CreateStockMovement(context.Background(), invalidMovement, items)
		assert.Error(t, err)

		// Test creating issue movement with insufficient stock
		issueMovement := &models.StockMovement{
			ReferenceNo:  "ISS001",
			MovementType: "issue",
			WarehouseID:  warehouse.ID,
			MovementDate: time.Now(),
			Status:       "pending",
			Notes:        "Test issue",
		}

		issueItems := []models.StockMovementItem{
			{
				ProductID:   product.ID,
				Quantity:    200, // More than available stock
				BatchNumber: "BATCH001",
			},
		}

		err = svc.CreateStockMovement(context.Background(), issueMovement, issueItems)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient stock")
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		// Create test warehouse and product
		warehouse := &models.Warehouse{
			Code:        "WH002",
			Name:        "Test Warehouse",
			Description: "Test Description",
			Address:     "123 Test St",
			IsActive:    true,
			Capacity:    1000.0,
		}
		err := svc.CreateWarehouse(context.Background(), warehouse)
		require.NoError(t, err)

		category := &models.ProductCategory{
			Code:        "CAT002",
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    nil,
			IsActive:    true,
		}
		err = svc.CreateProductCategory(context.Background(), category)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "PROD002",
			Name:          "Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			UnitPrice:     100.0,
			ReorderPoint:  10,
			MinimumOrder:  5,
			LeadTime:      7,
			IsActive:      true,
		}
		err = svc.CreateProduct(context.Background(), product)
		require.NoError(t, err)

		// Try to create a stock movement with invalid items to trigger rollback
		movement := &models.StockMovement{
			ReferenceNo:  "RCV003",
			MovementType: "receipt",
			WarehouseID:  warehouse.ID,
			MovementDate: time.Now(),
			Status:       "pending",
			Notes:        "Test rollback",
		}

		invalidItems := []models.StockMovementItem{
			{
				ProductID:   product.ID,
				Quantity:    -100, // Invalid negative quantity
				UnitCost:    90.0,
				BatchNumber: "BATCH001",
				ExpiryDate:  time.Now().AddDate(1, 0, 0),
			},
		}

		err = svc.CreateStockMovement(context.Background(), movement, invalidItems)
		assert.Error(t, err)

		// Verify no stock levels were affected
		var stockLevel float64
		err = testDB.DB.DB().QueryRowContext(context.Background(),
			"SELECT COALESCE(SUM(quantity), 0) FROM stock_levels WHERE product_id = ? AND warehouse_id = ?",
			product.ID, warehouse.ID).Scan(&stockLevel)
		require.NoError(t, err)
		assert.Equal(t, 0.0, stockLevel)
	})

	t.Run("ConcurrentStockMovements", func(t *testing.T) {
		// Create test warehouse and product
		warehouse := &models.Warehouse{
			Code:        "WH003",
			Name:        "Test Warehouse",
			Description: "Test Description",
			Address:     "123 Test St",
			IsActive:    true,
			Capacity:    1000.0,
		}
		err := svc.CreateWarehouse(context.Background(), warehouse)
		require.NoError(t, err)

		category := &models.ProductCategory{
			Code:        "CAT003",
			Name:        "Test Category",
			Description: "Test Description",
			ParentID:    nil,
			IsActive:    true,
		}
		err = svc.CreateProductCategory(context.Background(), category)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "PROD003",
			Name:          "Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			UnitPrice:     100.0,
			ReorderPoint:  10,
			MinimumOrder:  5,
			LeadTime:      7,
			IsActive:      true,
		}
		err = svc.CreateProduct(context.Background(), product)
		require.NoError(t, err)

		// Create initial stock receipt
		initMovement := &models.StockMovement{
			ReferenceNo:  "RCV004",
			MovementType: "receipt",
			WarehouseID:  warehouse.ID,
			MovementDate: time.Now(),
			Status:       "completed",
			Notes:        "Initial stock",
		}

		initItems := []models.StockMovementItem{
			{
				ProductID:   product.ID,
				Quantity:    1000,
				UnitCost:    90.0,
				BatchNumber: "BATCH001",
				ExpiryDate:  time.Now().AddDate(1, 0, 0),
			},
		}

		err = svc.CreateStockMovement(context.Background(), initMovement, initItems)
		require.NoError(t, err)

		// Create multiple issue movements concurrently
		done := make(chan bool)
		for i := 0; i < 5; i++ {
			go func(i int) {
				movement := &models.StockMovement{
					ReferenceNo:  fmt.Sprintf("ISS00%d", i+1),
					MovementType: "issue",
					WarehouseID:  warehouse.ID,
					MovementDate: time.Now(),
					Status:       "pending",
					Notes:        fmt.Sprintf("Concurrent issue %d", i+1),
				}

				items := []models.StockMovementItem{
					{
						ProductID:   product.ID,
						Quantity:    100,
						BatchNumber: "BATCH001",
					},
				}

				err := svc.CreateStockMovement(context.Background(), movement, items)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-done
		}

		// Verify final stock level is correct (1000 - 5*100 = 500)
		var finalStock float64
		err = testDB.DB.DB().QueryRowContext(context.Background(),
			"SELECT quantity FROM stock_levels WHERE product_id = ? AND warehouse_id = ?",
			product.ID, warehouse.ID).Scan(&finalStock)
		require.NoError(t, err)
		assert.Equal(t, 500.0, finalStock)
	})

	t.Run("GetProductStock", func(t *testing.T) {
		warehouse := &models.Warehouse{
			Code:        "WH-STOCK",
			Name:        "Stock Test Warehouse",
			Description: "Test Description",
			Address:     "123 Stock St",
			IsActive:    true,
			Capacity:    1000.0,
		}
		err := svc.CreateWarehouse(context.Background(), warehouse)
		require.NoError(t, err)

		product := &models.Product{
			Code:          "STOCK-TEST",
			Name:          "Stock Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			ReorderPoint:  10,
			IsActive:      true,
		}
		err = svc.CreateProduct(context.Background(), product)
		require.NoError(t, err)

		// Create an inbound transaction
		transaction := &models.InventoryTransaction{
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    50.0,
			Type:        "RECEIPT",
			Reference:   "TEST-RECEIPT-001",
		}
		err = svc.CreateInventoryTransaction(context.Background(), transaction)
		require.NoError(t, err)

		// Create an outbound transaction
		outbound := &models.InventoryTransaction{
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    -20.0,
			Type:        "ISSUE",
			Reference:   "TEST-ISSUE-001",
		}
		err = svc.CreateInventoryTransaction(context.Background(), outbound)
		require.NoError(t, err)

		// Get stock levels
		stock, err := svc.GetProductStock(context.Background(), product.ID)
		require.NoError(t, err)
		assert.Equal(t, 30.0, stock[warehouse.ID])
	})

	t.Run("GetLowStockProducts", func(t *testing.T) {
		// Create product with low stock threshold
		lowStockProduct := &models.Product{
			Code:          "LOW-STOCK",
			Name:          "Low Stock Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			ReorderPoint:  100,
			IsActive:      true,
		}
		err := svc.CreateProduct(context.Background(), lowStockProduct)
		require.NoError(t, err)

		// Add some stock but below reorder point
		transaction := &models.InventoryTransaction{
			ProductID:   lowStockProduct.ID,
			WarehouseID: warehouse.ID,
			Quantity:    50.0,
			Type:        "RECEIPT",
			Reference:   "TEST-LOW-STOCK",
		}
		err = svc.CreateInventoryTransaction(context.Background(), transaction)
		require.NoError(t, err)

		// Check low stock products
		lowStockProducts, err := svc.GetLowStockProducts(context.Background())
		require.NoError(t, err)
		found := false
		for _, p := range lowStockProducts {
			if p.ID == lowStockProduct.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Low stock product should be in the results")
	})

	t.Run("InventoryTransactionValidation", func(t *testing.T) {
		product := &models.Product{
			Code:          "INV-TEST",
			Name:          "Inventory Test Product",
			Description:   "Test Description",
			CategoryID:    category.ID,
			UnitOfMeasure: "PCS",
			ReorderPoint:  10,
			IsActive:      true,
		}
		err := svc.CreateProduct(context.Background(), product)
		require.NoError(t, err)

		// Try to create outbound transaction with insufficient stock
		invalidOutbound := &models.InventoryTransaction{
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    -50.0,
			Type:        "ISSUE",
			Reference:   "TEST-INVALID-ISSUE",
		}
		err = svc.CreateInventoryTransaction(context.Background(), invalidOutbound)
		assert.Error(t, err, "Should not allow outbound transaction with insufficient stock")

		// Add stock and then try valid outbound
		receipt := &models.InventoryTransaction{
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    100.0,
			Type:        "RECEIPT",
			Reference:   "TEST-RECEIPT",
		}
		err = svc.CreateInventoryTransaction(context.Background(), receipt)
		require.NoError(t, err)

		validOutbound := &models.InventoryTransaction{
			ProductID:   product.ID,
			WarehouseID: warehouse.ID,
			Quantity:    -50.0,
			Type:        "ISSUE",
			Reference:   "TEST-VALID-ISSUE",
		}
		err = svc.CreateInventoryTransaction(context.Background(), validOutbound)
		assert.NoError(t, err, "Should allow outbound transaction with sufficient stock")
	})
}
