package hr

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

func (s *Service) CreateDepartment(ctx context.Context, dept *models.Department) error {
	query := `
		INSERT INTO departments (
			code, name, parent_department_id, manager_id,
			cost_center, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		dept.Code,
		dept.Name,
		dept.ParentDepartmentID,
		dept.ManagerID,
		dept.CostCenter,
		dept.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating department: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	dept.ID = id
	return nil
}

func (s *Service) CreateJobPosition(ctx context.Context, position *models.JobPosition) error {
	query := `
		INSERT INTO job_positions (
			code, title, department_id, grade_level,
			salary_min, salary_max, requirements,
			is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		position.Code,
		position.Title,
		position.DepartmentID,
		position.GradeLevel,
		position.SalaryMin,
		position.SalaryMax,
		position.Requirements,
		position.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating job position: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	position.ID = id
	return nil
}

func (s *Service) CreateEmployee(ctx context.Context, employee *models.Employee) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Validate department exists
	var deptExists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM departments WHERE id = ? AND is_active = true)",
		employee.DepartmentID).Scan(&deptExists)
	if err != nil {
		return fmt.Errorf("error checking department: %v", err)
	}
	if !deptExists {
		return errors.New("invalid or inactive department")
	}

	// Validate job position exists and belongs to department
	var positionExists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM job_positions 
			WHERE id = ? AND department_id = ? AND is_active = true
		)`, employee.JobPositionID, employee.DepartmentID).Scan(&positionExists)
	if err != nil {
		return fmt.Errorf("error checking job position: %v", err)
	}
	if !positionExists {
		return errors.New("invalid or inactive job position for department")
	}

	query := `
		INSERT INTO employees (
			employee_number, first_name, last_name,
			email, phone, hire_date, department_id,
			job_position_id, manager_id, status,
			employment_type, tax_id, bank_account,
			address, emergency_contact,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		employee.EmployeeNumber,
		employee.FirstName,
		employee.LastName,
		employee.Email,
		employee.Phone,
		employee.HireDate,
		employee.DepartmentID,
		employee.JobPositionID,
		employee.ManagerID,
		employee.Status,
		employee.EmploymentType,
		employee.TaxID,
		employee.BankAccount,
		employee.Address,
		employee.EmergencyContact,
	)
	if err != nil {
		return fmt.Errorf("error creating employee: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	employee.ID = id

	return tx.Commit()
}

func (s *Service) CreateSalaryComponent(ctx context.Context, component *models.SalaryComponent) error {
	query := `
		INSERT INTO salary_components (
			code, name, type, is_taxable,
			calculation_rule, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		component.Code,
		component.Name,
		component.Type,
		component.IsTaxable,
		component.CalculationRule,
		component.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating salary component: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	component.ID = id
	return nil
}

func (s *Service) AssignEmployeeSalary(ctx context.Context, salary *models.EmployeeSalary) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Close any existing active salary components
	query := `
		UPDATE employee_salary 
		SET effective_to = ? 
		WHERE employee_id = ? 
		AND component_id = ?
		AND (effective_to IS NULL OR effective_to > ?)
	`
	_, err = tx.ExecContext(ctx, query,
		salary.EffectiveFrom,
		salary.EmployeeID,
		salary.ComponentID,
		salary.EffectiveFrom,
	)
	if err != nil {
		return fmt.Errorf("error updating existing salary components: %v", err)
	}

	// Insert new salary component
	query = `
		INSERT INTO employee_salary (
			employee_id, component_id, amount,
			currency_code, effective_from, effective_to,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		salary.EmployeeID,
		salary.ComponentID,
		salary.Amount,
		salary.CurrencyCode,
		salary.EffectiveFrom,
		salary.EffectiveTo,
	)
	if err != nil {
		return fmt.Errorf("error creating employee salary: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	salary.ID = id

	return tx.Commit()
}

func (s *Service) RecordAttendance(ctx context.Context, record *models.AttendanceRecord) error {
	query := `
		INSERT INTO attendance_records (
			employee_id, date, check_in,
			check_out, status, notes,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		record.EmployeeID,
		record.Date,
		record.CheckIn,
		record.CheckOut,
		record.Status,
		record.Notes,
	)
	if err != nil {
		return fmt.Errorf("error creating attendance record: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	record.ID = id
	return nil
}

func (s *Service) CreateLeaveType(ctx context.Context, leaveType *models.LeaveType) error {
	query := `
		INSERT INTO leave_types (
			code, name, description, paid,
			annual_allowance, is_active,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := s.db.DB().ExecContext(ctx, query,
		leaveType.Code,
		leaveType.Name,
		leaveType.Description,
		leaveType.Paid,
		leaveType.AnnualAllowance,
		leaveType.IsActive,
	)
	if err != nil {
		return fmt.Errorf("error creating leave type: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	leaveType.ID = id
	return nil
}

func (s *Service) RequestLeave(ctx context.Context, request *models.LeaveRequest) error {
	tx, err := s.db.DB().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	// Check leave balance
	var leaveAllowance, usedLeave float64
	err = tx.QueryRowContext(ctx, "SELECT annual_allowance FROM leave_types WHERE id = ?",
		request.LeaveTypeID).Scan(&leaveAllowance)
	if err != nil {
		return fmt.Errorf("error checking leave allowance: %v", err)
	}

	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_days), 0)
		FROM leave_requests
		WHERE employee_id = ?
		AND leave_type_id = ?
		AND status != 'rejected'
		AND YEAR(start_date) = YEAR(?)
	`, request.EmployeeID, request.LeaveTypeID, request.StartDate).Scan(&usedLeave)
	if err != nil {
		return fmt.Errorf("error checking used leave: %v", err)
	}

	if usedLeave+request.TotalDays > leaveAllowance {
		return errors.New("insufficient leave balance")
	}

	// Check for overlapping leave requests
	var overlapping int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM leave_requests
		WHERE employee_id = ?
		AND status != 'rejected'
		AND (
			(start_date <= ? AND end_date >= ?) OR
			(start_date <= ? AND end_date >= ?) OR
			(start_date >= ? AND end_date <= ?)
		)
	`, request.EmployeeID, request.StartDate, request.StartDate,
		request.EndDate, request.EndDate,
		request.StartDate, request.EndDate).Scan(&overlapping)
	if err != nil {
		return fmt.Errorf("error checking overlapping leaves: %v", err)
	}

	if overlapping > 0 {
		return errors.New("overlapping leave request exists")
	}

	query := `
		INSERT INTO leave_requests (
			employee_id, leave_type_id, start_date,
			end_date, total_days, status, reason,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := tx.ExecContext(ctx, query,
		request.EmployeeID,
		request.LeaveTypeID,
		request.StartDate,
		request.EndDate,
		request.TotalDays,
		request.Status,
		request.Reason,
	)
	if err != nil {
		return fmt.Errorf("error creating leave request: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %v", err)
	}
	request.ID = id

	return tx.Commit()
}

func (s *Service) GetEmployeeLeaveBalance(ctx context.Context, employeeID int64, leaveTypeID int64) (float64, error) {
	var allowance, used float64

	err := s.db.DB().QueryRowContext(ctx, "SELECT annual_allowance FROM leave_types WHERE id = ?",
		leaveTypeID).Scan(&allowance)
	if err != nil {
		return 0, fmt.Errorf("error getting leave allowance: %v", err)
	}

	err = s.db.DB().QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_days), 0)
		FROM leave_requests
		WHERE employee_id = ?
		AND leave_type_id = ?
		AND status != 'rejected'
		AND YEAR(start_date) = YEAR(CURRENT_DATE())
	`, employeeID, leaveTypeID).Scan(&used)
	if err != nil {
		return 0, fmt.Errorf("error getting used leave: %v", err)
	}

	return allowance - used, nil
}
