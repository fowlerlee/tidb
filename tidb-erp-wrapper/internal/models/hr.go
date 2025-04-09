package models

import "time"

type Department struct {
	ID                 int64     `json:"id"`
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	ParentDepartmentID *int64    `json:"parent_department_id,omitempty"`
	ManagerID          *int64    `json:"manager_id,omitempty"`
	CostCenter         string    `json:"cost_center"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type JobPosition struct {
	ID           int64     `json:"id"`
	Code         string    `json:"code"`
	Title        string    `json:"title"`
	DepartmentID int64     `json:"department_id"`
	GradeLevel   string    `json:"grade_level"`
	SalaryMin    float64   `json:"salary_min"`
	SalaryMax    float64   `json:"salary_max"`
	Requirements string    `json:"requirements"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Employee struct {
	ID               int64     `json:"id"`
	EmployeeNumber   string    `json:"employee_number"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	HireDate         time.Time `json:"hire_date"`
	DepartmentID     int64     `json:"department_id"`
	JobPositionID    int64     `json:"job_position_id"`
	ManagerID        *int64    `json:"manager_id,omitempty"`
	Status           string    `json:"status"`
	EmploymentType   string    `json:"employment_type"`
	TaxID            string    `json:"tax_id"`
	BankAccount      string    `json:"bank_account"`
	Address          string    `json:"address"`
	EmergencyContact string    `json:"emergency_contact"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type SalaryComponent struct {
	ID              int64     `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	IsTaxable       bool      `json:"is_taxable"`
	CalculationRule string    `json:"calculation_rule"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type EmployeeSalary struct {
	ID            int64      `json:"id"`
	EmployeeID    int64      `json:"employee_id"`
	ComponentID   int64      `json:"component_id"`
	Amount        float64    `json:"amount"`
	CurrencyCode  string     `json:"currency_code"`
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AttendanceRecord struct {
	ID         int64      `json:"id"`
	EmployeeID int64      `json:"employee_id"`
	Date       time.Time  `json:"date"`
	CheckIn    *time.Time `json:"check_in,omitempty"`
	CheckOut   *time.Time `json:"check_out,omitempty"`
	Status     string     `json:"status"`
	Notes      string     `json:"notes"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type LeaveType struct {
	ID              int64     `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Paid            bool      `json:"paid"`
	AnnualAllowance float64   `json:"annual_allowance"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type LeaveRequest struct {
	ID          int64      `json:"id"`
	EmployeeID  int64      `json:"employee_id"`
	LeaveTypeID int64      `json:"leave_type_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	TotalDays   float64    `json:"total_days"`
	Status      string     `json:"status"`
	Reason      string     `json:"reason"`
	ApprovedBy  *int64     `json:"approved_by,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
