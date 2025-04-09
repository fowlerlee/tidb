package quality

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/db"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
)

type Service struct {
	db *db.DBHandler
}

func NewService(db *db.DBHandler) *Service {
	return &Service{db: db}
}

func (s *Service) CreateQualityParameter(ctx context.Context, param *models.QualityParameter) error {
	query := `
		INSERT INTO quality_parameters (
			code, name, description, measurement_unit,
			min_value, max_value, target_value,
			is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		param.Code,
		param.Name,
		param.Description,
		param.MeasurementUnit,
		param.MinValue,
		param.MaxValue,
		param.TargetValue,
		param.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating quality parameter: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	param.ID = id
	return nil
}

func (s *Service) CreateInspectionPoint(ctx context.Context, point *models.InspectionPoint) error {
	query := `
		INSERT INTO inspection_points (
			code, name, description, process_type,
			department_id, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		point.Code,
		point.Name,
		point.Description,
		point.ProcessType,
		point.DepartmentID,
		point.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating inspection point: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	point.ID = id
	return nil
}

func (s *Service) CreateInspectionPlan(ctx context.Context, plan *models.InspectionPlan, parameters []models.InspectionPlanParameter) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO inspection_plans (
			code, name, description, product_id,
			version, status, effective_from, effective_to,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		plan.Code,
		plan.Name,
		plan.Description,
		plan.ProductID,
		plan.Version,
		plan.Status,
		plan.EffectiveFrom,
		plan.EffectiveTo,
	)
	if err != nil {
		return fmt.Errorf("error creating inspection plan: %v", err)
	}

	planID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting plan ID: %v", err)
	}
	plan.ID = planID

	// Insert inspection plan parameters
	for _, param := range parameters {
		query := `
			INSERT INTO inspection_plan_parameters (
				inspection_plan_id, inspection_point_id,
				parameter_id, sampling_method, sample_size,
				mandatory, sequence_no,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		`
		_, err := tx.ExecContext(ctx, query,
			planID,
			param.InspectionPointID,
			param.ParameterID,
			param.SamplingMethod,
			param.SampleSize,
			param.Mandatory,
			param.SequenceNo,
		)
		if err != nil {
			return fmt.Errorf("error creating inspection plan parameter: %v", err)
		}
	}

	return tx.Commit()
}

func (s *Service) CreateQualityInspection(ctx context.Context, inspection *models.QualityInspection, results []models.InspectionResult) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Check if inspection plan is active
	var planStatus string
	err = tx.QueryRowContext(ctx, "SELECT status FROM inspection_plans WHERE id = ?",
		inspection.InspectionPlanID).Scan(&planStatus)
	if err != nil {
		return fmt.Errorf("error checking inspection plan: %v", err)
	}
	if planStatus != "active" {
		return errors.New("inspection plan is not active")
	}

	// Insert quality inspection
	query := `
		INSERT INTO quality_inspections (
			inspection_number, inspection_plan_id,
			reference_type, reference_id, inspector_id,
			inspection_date, status, result, notes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		inspection.InspectionNumber,
		inspection.InspectionPlanID,
		inspection.ReferenceType,
		inspection.ReferenceID,
		inspection.InspectorID,
		inspection.InspectionDate,
		inspection.Status,
		inspection.Result,
		inspection.Notes,
	)
	if err != nil {
		return fmt.Errorf("error creating quality inspection: %v", err)
	}

	inspectionID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting inspection ID: %v", err)
	}
	inspection.ID = inspectionID

	// Insert inspection results
	for _, res := range results {
		// Validate parameter against plan
		var exists bool
		err = tx.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM inspection_plan_parameters
				WHERE inspection_plan_id = ?
				AND parameter_id = ?
			)`, inspection.InspectionPlanID, res.ParameterID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("error validating parameter: %v", err)
		}
		if !exists {
			return fmt.Errorf("parameter %d is not part of the inspection plan", res.ParameterID)
		}

		// Get parameter details for validation
		var minValue, maxValue sql.NullFloat64
		err = tx.QueryRowContext(ctx, `
			SELECT min_value, max_value
			FROM quality_parameters
			WHERE id = ?
		`, res.ParameterID).Scan(&minValue, &maxValue)
		if err != nil {
			return fmt.Errorf("error getting parameter details: %v", err)
		}

		// Validate measurement against limits
		if res.MeasuredValue != nil {
			if minValue.Valid && *res.MeasuredValue < minValue.Float64 {
				res.IsConforming = false
			}
			if maxValue.Valid && *res.MeasuredValue > maxValue.Float64 {
				res.IsConforming = false
			}
		}

		query := `
			INSERT INTO inspection_results (
				inspection_id, parameter_id,
				measured_value, is_conforming, notes,
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, NOW(), NOW())
		`
		_, err := tx.ExecContext(ctx, query,
			inspectionID,
			res.ParameterID,
			res.MeasuredValue,
			res.IsConforming,
			res.Notes,
		)
		if err != nil {
			return fmt.Errorf("error creating inspection result: %v", err)
		}
	}

	// Update overall inspection result based on results
	var nonConforming int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM inspection_results
		WHERE inspection_id = ?
		AND is_conforming = false
	`, inspectionID).Scan(&nonConforming)
	if err != nil {
		return fmt.Errorf("error counting non-conforming results: %v", err)
	}

	if nonConforming > 0 {
		_, err = tx.ExecContext(ctx, `
			UPDATE quality_inspections 
			SET result = 'non-conforming'
			WHERE id = ?
		`, inspectionID)
		if err != nil {
			return fmt.Errorf("error updating inspection result: %v", err)
		}
		inspection.Result = "non-conforming"
	}

	return tx.Commit()
}

func (s *Service) CreateQualityNotification(ctx context.Context, notification *models.QualityNotification) error {
	query := `
		INSERT INTO quality_notifications (
			notification_number, type, priority,
			reference_type, reference_id, reported_by,
			assigned_to, status, description,
			root_cause, corrective_action, due_date,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		notification.NotificationNumber,
		notification.Type,
		notification.Priority,
		notification.ReferenceType,
		notification.ReferenceID,
		notification.ReportedBy,
		notification.AssignedTo,
		notification.Status,
		notification.Description,
		notification.RootCause,
		notification.CorrectiveAction,
		notification.DueDate,
	)
	if err != nil {
		return fmt.Errorf("error creating quality notification: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	notification.ID = id
	return nil
}

func (s *Service) GetInspectionPlanByID(ctx context.Context, id int64) (*models.InspectionPlan, []models.InspectionPlanParameter, error) {
	plan := &models.InspectionPlan{}
	query := `
		SELECT id, code, name, description,
			   product_id, version, status,
			   effective_from, effective_to,
			   created_at, updated_at
		FROM inspection_plans
		WHERE id = ?
	`
	err := s.db.DB().QueryRowContext(ctx, query, id).Scan(
		&plan.ID,
		&plan.Code,
		&plan.Name,
		&plan.Description,
		&plan.ProductID,
		&plan.Version,
		&plan.Status,
		&plan.EffectiveFrom,
		&plan.EffectiveTo,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("inspection plan not found")
		}
		return nil, nil, fmt.Errorf("error querying inspection plan: %v", err)
	}

	parameters := []models.InspectionPlanParameter{}
	query = `
		SELECT id, inspection_point_id, parameter_id,
			   sampling_method, sample_size, mandatory,
			   sequence_no, created_at, updated_at
		FROM inspection_plan_parameters
		WHERE inspection_plan_id = ?
		ORDER BY sequence_no
	`
	rows, err := s.db.DB().QueryContext(ctx, query, id)
	if err != nil {
		return nil, nil, fmt.Errorf("error querying plan parameters: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var param models.InspectionPlanParameter
		err := rows.Scan(
			&param.ID,
			&param.InspectionPointID,
			&param.ParameterID,
			&param.SamplingMethod,
			&param.SampleSize,
			&param.Mandatory,
			&param.SequenceNo,
			&param.CreatedAt,
			&param.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning plan parameter: %v", err)
		}
		param.InspectionPlanID = id
		parameters = append(parameters, param)
	}

	return plan, parameters, nil
}
