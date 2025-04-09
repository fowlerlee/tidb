package quality

import (
	"context"
	"testing"
	"tidb-erp-wrapper/internal/models"
	"tidb-erp-wrapper/internal/testutil"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateQualityParameter(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer db.Close()

	svc := NewService(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		param   *models.QualityParameter
		wantErr bool
	}{
		{
			name: "valid parameter",
			param: &models.QualityParameter{
				Code:            "TEMP-001",
				Name:            "Temperature",
				Description:     "Temperature measurement",
				MeasurementUnit: "°C",
				MinValue:        0,
				MaxValue:        100,
				TargetValue:     25,
				IsActive:        true,
			},
			wantErr: false,
		},
		{
			name: "duplicate code",
			param: &models.QualityParameter{
				Code:            "TEMP-001",
				Name:            "Temperature 2",
				MeasurementUnit: "°C",
				IsActive:        true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.CreateQualityParameter(ctx, tt.param)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.param.ID)
			}
		})
	}
}

func TestCreateInspectionPlan(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer db.Close()

	svc := NewService(db)
	ctx := context.Background()

	// Create test parameter
	param := &models.QualityParameter{
		Code:            "TEST-001",
		Name:            "Test Parameter",
		MeasurementUnit: "units",
		IsActive:        true,
	}
	err := svc.CreateQualityParameter(ctx, param)
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

	now := time.Now()
	tests := []struct {
		name       string
		plan       *models.InspectionPlan
		parameters []models.InspectionPlanParameter
		wantErr    bool
	}{
		{
			name: "valid plan",
			plan: &models.InspectionPlan{
				Code:          "PLAN-001",
				Name:          "Test Plan",
				ProductID:     1,
				Version:       "1.0",
				Status:        "active",
				EffectiveFrom: now,
				EffectiveTo:   now.AddDate(1, 0, 0),
			},
			parameters: []models.InspectionPlanParameter{
				{
					InspectionPointID: point.ID,
					ParameterID:       param.ID,
					SamplingMethod:    "random",
					SampleSize:        5,
					Mandatory:         true,
					SequenceNo:        1,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.CreateInspectionPlan(ctx, tt.plan, tt.parameters)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.plan.ID)

				// Verify we can retrieve the plan
				retrievedPlan, retrievedParams, err := svc.GetInspectionPlanByID(ctx, tt.plan.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.plan.Code, retrievedPlan.Code)
				assert.Equal(t, tt.plan.Name, retrievedPlan.Name)
				assert.Len(t, retrievedParams, len(tt.parameters))
			}
		})
	}
}

func TestCreateQualityInspection(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer db.Close()

	svc := NewService(db)
	ctx := context.Background()

	// Create test parameter
	param := &models.QualityParameter{
		Code:            "TEST-002",
		Name:            "Test Parameter",
		MeasurementUnit: "units",
		MinValue:        0,
		MaxValue:        100,
		IsActive:        true,
	}
	err := svc.CreateQualityParameter(ctx, param)
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
	now := time.Now()
	plan := &models.InspectionPlan{
		Code:          "PLAN-002",
		Name:          "Test Plan",
		ProductID:     1,
		Version:       "1.0",
		Status:        "active",
		EffectiveFrom: now,
		EffectiveTo:   now.AddDate(1, 0, 0),
	}
	planParams := []models.InspectionPlanParameter{
		{
			InspectionPointID: point.ID,
			ParameterID:       param.ID,
			SamplingMethod:    "random",
			SampleSize:        5,
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
}

func TestCreateQualityNotification(t *testing.T) {
	db := testutil.NewTestDB(t)
	defer db.Close()

	svc := NewService(db)
	ctx := context.Background()

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
				AssignedTo:         2,
				Status:             "open",
				Description:        "Quality issue found",
				DueDate:            time.Now().AddDate(0, 0, 7),
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
}
