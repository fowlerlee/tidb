package quality

import (
	"context"
	"testing"
	"time"

	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQualityService(t *testing.T) {
	testDB, err := testutil.NewTestDB()
	require.NoError(t, err)
	defer testDB.Cleanup()

	svc := NewService(testDB.DB)
	ctx := context.Background()

	// Create test parameter first
	minVal, maxVal := float64(0), float64(100)
	param := &models.QualityParameter{
		Code:            "PARAM-001",
		Name:            "Test Parameter",
		Description:     "Test parameter for inspection",
		MeasurementUnit: "units",
		MinValue:        &minVal,
		MaxValue:        &maxVal,
		IsActive:        true,
	}
	err = svc.CreateQualityParameter(ctx, param)
	require.NoError(t, err)

	// Create test inspection point
	point := &models.InspectionPoint{
		Code:        "IP-001",
		Name:        "Test Point",
		ProcessType: "manufacturing",
		IsActive:    true,
	}
	err = svc.CreateInspectionPoint(ctx, point)
	require.NoError(t, err)

	t.Run("CreateQualityParameter", func(t *testing.T) {
		// Create pointers to numeric values
		min := float64(0)
		max := float64(100)
		target := float64(25)

		newParam := &models.QualityParameter{
			Code:            "TEMP-001",
			Name:            "Temperature",
			Description:     "Temperature measurement",
			MeasurementUnit: "°C",
			MinValue:        &min,
			MaxValue:        &max,
			TargetValue:     &target,
			IsActive:        true,
		}

		err := svc.CreateQualityParameter(ctx, newParam)
		assert.NoError(t, err)
		assert.NotZero(t, newParam.ID)
	})

	t.Run("CreateInspectionPlan", func(t *testing.T) {
		now := time.Now()
		effectiveTo := now.AddDate(1, 0, 0)
		sampleSize := 5
		var productID int64 = 1

		plan := &models.InspectionPlan{
			Code:          "PLAN-001",
			Name:          "Test Plan",
			ProductID:     &productID,
			Version:       "1.0",
			Status:        "active",
			EffectiveFrom: now,
			EffectiveTo:   &effectiveTo,
		}

		params := []models.InspectionPlanParameter{
			{
				InspectionPointID: point.ID,
				ParameterID:       param.ID,
				SamplingMethod:    "random",
				SampleSize:        &sampleSize,
				Mandatory:         true,
				SequenceNo:        1,
			},
		}

		err = svc.CreateInspectionPlan(ctx, plan, params)
		assert.NoError(t, err)
		assert.NotZero(t, plan.ID)
	})

	t.Run("CreateQualityInspection", func(t *testing.T) {
		now := time.Now()
		effectiveTo := now.AddDate(1, 0, 0)
		var productID int64 = 1
		sampleSize := 5

		// Create test parameter
		param := &models.QualityParameter{
			Code:            "TEST-002",
			Name:            "Test Parameter",
			MeasurementUnit: "units",
			MinValue:        &minVal,
			MaxValue:        &maxVal,
			IsActive:        true,
		}
		err = svc.CreateQualityParameter(ctx, param)
		require.NoError(t, err)

		// Create test inspection point
		point := &models.InspectionPoint{
			Code:        "IP-002",
			Name:        "Test Point",
			ProcessType: "manufacturing",
			IsActive:    true,
		}
		err = svc.CreateInspectionPoint(ctx, point)
		require.NoError(t, err)

		// Create test inspection plan
		plan := &models.InspectionPlan{
			Code:          "PLAN-002",
			Name:          "Test Plan",
			ProductID:     &productID,
			Version:       "1.0",
			Status:        "active",
			EffectiveFrom: now,
			EffectiveTo:   &effectiveTo,
		}
		planParams := []models.InspectionPlanParameter{
			{
				InspectionPointID: point.ID,
				ParameterID:       param.ID,
				SamplingMethod:    "random",
				SampleSize:        &sampleSize,
				Mandatory:         true,
				SequenceNo:        1,
			},
		}
		err = svc.CreateInspectionPlan(ctx, plan, planParams)
		require.NoError(t, err)

		measuredValue := 50.0
		tests := []struct {
			name       string
			inspection *models.QualityInspection
			results    []models.InspectionResult
			wantErr    bool
		}{
			{
				name: "valid inspection - conforming",
				inspection: &models.QualityInspection{
					InspectionNumber: "INSP-001",
					InspectionPlanID: plan.ID,
					ReferenceType:    "production_order",
					ReferenceID:      1,
					InspectorID:      1,
					InspectionDate:   now,
					Status:           "completed",
					Result:           "conforming",
				},
				results: []models.InspectionResult{
					{
						ParameterID:   param.ID,
						MeasuredValue: &measuredValue,
						IsConforming:  true,
					},
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := svc.CreateQualityInspection(ctx, tt.inspection, tt.results)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotZero(t, tt.inspection.ID)
				}
			})
		}
	})

	t.Run("CreateQualityNotification", func(t *testing.T) {
		assignedTo := int64(2)
		dueDate := time.Now().AddDate(0, 0, 7)

		tests := []struct {
			name         string
			notification *models.QualityNotification
			wantErr      bool
		}{
			{
				name: "valid notification",
				notification: &models.QualityNotification{
					NotificationNumber: "QN-001",
					Type:               "non_conformance",
					Priority:           "high",
					ReferenceType:      "inspection",
					ReferenceID:        1,
					ReportedBy:         1,
					AssignedTo:         &assignedTo,
					Status:             "open",
					Description:        "Quality issue found",
					DueDate:            &dueDate,
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := svc.CreateQualityNotification(ctx, tt.notification)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotZero(t, tt.notification.ID)
				}
			})
		}
	})
}
