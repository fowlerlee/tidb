package procurement

import (
	"context"
	"testing"

	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcurementService(t *testing.T) {
	testDB, err := testutil.NewTestDB()
	require.NoError(t, err)
	defer testDB.Cleanup()

	svc := NewService(testDB.DB)
	ctx := context.Background()

	t.Run("CreateSupplier", func(t *testing.T) {
		// Test creating a valid supplier
		supplier := testutil.GenerateSupplier(0)
		err := svc.CreateSupplier(ctx, supplier)
		assert.NoError(t, err)
		assert.NotZero(t, supplier.ID)

		// Test creating duplicate supplier code
		dupSupplier := testutil.GenerateSupplier(0)
		dupSupplier.Code = supplier.Code
		err = svc.CreateSupplier(ctx, dupSupplier)
		assert.Error(t, err)
	})

	t.Run("CreatePurchaseOrder", func(t *testing.T) {
		
		// Create test supplier first
		supplier := testutil.GenerateSupplier(0)
		err := svc.CreateSupplier(ctx, supplier)
		require.NoError(t, err)

		// Test creating a valid purchase order
		po := testutil.GeneratePurchaseOrder(0, supplier.ID)
		items := []*models.PurchaseOrderItem{
			testutil.GeneratePurchaseOrderItem(0, 0),
			testutil.GeneratePurchaseOrderItem(0, 0),
		}

		// Convert []*models.PurchaseOrderItem to []models.PurchaseOrderItem
		convertedItems := make([]models.PurchaseOrderItem, len(items))
		for i, item := range items {
			convertedItems[i] = *item
		}
		err = svc.CreatePurchaseOrder(ctx, po, convertedItems)
		assert.NoError(t, err)
		assert.NotZero(t, po.ID)

		// Test creating order with invalid supplier
		invalidPO := testutil.GeneratePurchaseOrder(0, 999999)
		err = svc.CreatePurchaseOrder(ctx, invalidPO, items)
		assert.Error(t, err)

		// Test creating order with empty items
		emptyPO := testutil.GeneratePurchaseOrder(0, supplier.ID)
		err = svc.CreatePurchaseOrder(ctx, emptyPO, nil)
		assert.Error(t, err)
	})

	t.Run("GetPurchaseOrderByID", func(t *testing.T) {
		// Create test data
		supplier := testutil.GenerateSupplier(0)
		err := svc.CreateSupplier(ctx, supplier)
		require.NoError(t, err)

		po := testutil.GeneratePurchaseOrder(0, supplier.ID)
		items := []*models.PurchaseOrderItem{
			testutil.GeneratePurchaseOrderItem(0, 0),
		}
		err = svc.CreatePurchaseOrder(ctx, po, items)
		require.NoError(t, err)

		// Test retrieving existing order
		fetchedPO, fetchedItems, err := svc.GetPurchaseOrderByID(ctx, po.ID)
		assert.NoError(t, err)
		assert.NotNil(t, fetchedPO)
		assert.Equal(t, po.OrderNumber, fetchedPO.OrderNumber)
		assert.Len(t, fetchedItems, 1)

		// Test retrieving non-existent order
		fetchedPO, fetchedItems, err = svc.GetPurchaseOrderByID(ctx, 999999)
		assert.Error(t, err)
		assert.Nil(t, fetchedPO)
		assert.Nil(t, fetchedItems)
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		supplier := testutil.GenerateSupplier(0)
		err := svc.CreateSupplier(ctx, supplier)
		require.NoError(t, err)

		// Create a purchase order with invalid items to trigger rollback
		po := testutil.GeneratePurchaseOrder(0, supplier.ID)
		items := []*models.PurchaseOrderItem{
			{
				PurchaseOrderID: 0,
				ProductCode:     "", // Invalid product code
				Quantity:        -1, // Invalid quantity
			},
		}

		err = svc.CreatePurchaseOrder(ctx, po, items)
		assert.Error(t, err)

		// Verify the purchase order was not created
		fetchedPO, _, err := svc.GetPurchaseOrderByID(ctx, po.ID)
		assert.Error(t, err)
		assert.Nil(t, fetchedPO)
	})
}
