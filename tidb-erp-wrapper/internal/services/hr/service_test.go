package hr

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/models"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHRService(t *testing.T) {
	testDB, err := testutil.NewTestDB()
	require.NoError(t, err)
	defer testDB.Cleanup()

	svc := NewService(testDB.DB)

	t.Run("CreateDepartment", func(t *testing.T) {
		// Test creating a valid department
		dept := testutil.GenerateDepartment(0)
		err := svc.CreateDepartment(context.Background(), dept)
		assert.NoError(t, err)
		assert.NotZero(t, dept.ID)

		// Test creating duplicate department code
		dupDept := testutil.GenerateDepartment(0)
		dupDept.Code = dept.Code
		err = svc.CreateDepartment(context.Background(), dupDept)
		assert.Error(t, err)

		// Test creating department with invalid parent
		invalidParentDept := testutil.GenerateDepartment(0)
		invalidParentDept.ParentDepartmentID = new(int64)
		*invalidParentDept.ParentDepartmentID = 999999
		err = svc.CreateDepartment(context.Background(), invalidParentDept)
		assert.Error(t, err)
	})

	t.Run("CreateJobPosition", func(t *testing.T) {
		// Create test department first
		dept := testutil.GenerateDepartment(0)
		err := svc.CreateDepartment(context.Background(), dept)
		require.NoError(t, err)

		// Test creating a valid job position
		position := &models.JobPosition{
			Code:         "JP001",
			Title:        "Software Engineer",
			DepartmentID: dept.ID,
			GradeLevel:   "L4",
			SalaryMin:    50000.0,
			SalaryMax:    100000.0,
			Requirements: "5 years experience",
			IsActive:     true,
		}

		err = svc.CreateJobPosition(context.Background(), position)
		assert.NoError(t, err)
		assert.NotZero(t, position.ID)

		// Test creating position with invalid salary range
		invalidPosition := &models.JobPosition{
			Code:         "JP002",
			Title:        "Invalid Position",
			DepartmentID: dept.ID,
			GradeLevel:   "L4",
			SalaryMin:    100000.0,
			SalaryMax:    50000.0, // Max less than min
			Requirements: "5 years experience",
			IsActive:     true,
		}

		err = svc.CreateJobPosition(context.Background(), invalidPosition)
		assert.Error(t, err)

		// Test creating position with invalid department
		invalidDeptPosition := &models.JobPosition{
			Code:         "JP003",
			Title:        "Test Position",
			DepartmentID: 999999,
			GradeLevel:   "L4",
			SalaryMin:    50000.0,
			SalaryMax:    100000.0,
			Requirements: "5 years experience",
			IsActive:     true,
		}

		err = svc.CreateJobPosition(context.Background(), invalidDeptPosition)
		assert.Error(t, err)
	})

	t.Run("CreateEmployee", func(t *testing.T) {
		// Create test department and position
		dept := testutil.GenerateDepartment(0)
		err := svc.CreateDepartment(context.Background(), dept)
		require.NoError(t, err)

		position := &models.JobPosition{
			Code:         "JP001",
			Title:        "Software Engineer",
			DepartmentID: dept.ID,
			GradeLevel:   "L4",
			SalaryMin:    50000.0,
			SalaryMax:    100000.0,
			Requirements: "5 years experience",
			IsActive:     true,
		}
		err = svc.CreateJobPosition(context.Background(), position)
		require.NoError(t, err)

		// Test creating a valid employee
		employee := testutil.GenerateEmployee(0)
		employee.DepartmentID = dept.ID
		employee.JobPositionID = position.ID
		err = svc.CreateEmployee(context.Background(), employee)
		assert.NoError(t, err)
		assert.NotZero(t, employee.ID)

		// Test creating employee with duplicate employee number
		dupEmployee := testutil.GenerateEmployee(0)
		dupEmployee.EmployeeNumber = employee.EmployeeNumber
		dupEmployee.DepartmentID = dept.ID
		dupEmployee.JobPositionID = position.ID
		err = svc.CreateEmployee(context.Background(), dupEmployee)
		assert.Error(t, err)

		// Test creating employee with invalid department
		invalidDeptEmployee := testutil.GenerateEmployee(0)
		invalidDeptEmployee.DepartmentID = 999999
		invalidDeptEmployee.JobPositionID = position.ID
		err = svc.CreateEmployee(context.Background(), invalidDeptEmployee)
		assert.Error(t, err)

		// Test creating employee with invalid job position
		invalidPosEmployee := testutil.GenerateEmployee(0)
		invalidPosEmployee.DepartmentID = dept.ID
		invalidPosEmployee.JobPositionID = 999999
		err = svc.CreateEmployee(context.Background(), invalidPosEmployee)
		assert.Error(t, err)
	})

	t.Run("CreateLeaveRequest", func(t *testing.T) {
		// Create test employee
		dept := testutil.GenerateDepartment(0)
		err := svc.CreateDepartment(context.Background(), dept)
		require.NoError(t, err)

		position := &models.JobPosition{
			Code:         "JP001",
			Title:        "Software Engineer",
			DepartmentID: dept.ID,
			GradeLevel:   "L4",
			SalaryMin:    50000.0,
			SalaryMax:    100000.0,
			Requirements: "5 years experience",
			IsActive:     true,
		}
		err = svc.CreateJobPosition(context.Background(), position)
		require.NoError(t, err)

		employee := testutil.GenerateEmployee(0)
		employee.DepartmentID = dept.ID
		employee.JobPositionID = position.ID
		err = svc.CreateEmployee(context.Background(), employee)
		require.NoError(t, err)

		// Create leave type
		leaveType := &models.LeaveType{
			Code:            "AL",
			Name:            "Annual Leave",
			Description:     "Regular annual leave",
			Paid:            true,
			AnnualAllowance: 20.0,
			IsActive:        true,
		}
		err = svc.CreateLeaveType(context.Background(), leaveType)
		require.NoError(t, err)

		// Test creating a valid leave request
		request := &models.LeaveRequest{
			EmployeeID:  employee.ID,
			LeaveTypeID: leaveType.ID,
			StartDate:   time.Now().AddDate(0, 0, 1),
			EndDate:     time.Now().AddDate(0, 0, 5),
			TotalDays:   5.0,
			Status:      "pending",
			Reason:      "Vacation",
		}

		err = svc.RequestLeave(context.Background(), request)
		assert.NoError(t, err)
		assert.NotZero(t, request.ID)

		// Test creating leave request exceeding allowance
		excessRequest := &models.LeaveRequest{
			EmployeeID:  employee.ID,
			LeaveTypeID: leaveType.ID,
			StartDate:   time.Now().AddDate(0, 0, 10),
			EndDate:     time.Now().AddDate(0, 0, 40),
			TotalDays:   30.0, // Exceeds annual allowance
			Status:      "pending",
			Reason:      "Extended vacation",
		}

		err = svc.RequestLeave(context.Background(), excessRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient leave balance")

		// Test creating overlapping leave request
		overlapRequest := &models.LeaveRequest{
			EmployeeID:  employee.ID,
			LeaveTypeID: leaveType.ID,
			StartDate:   time.Now().AddDate(0, 0, 3),
			EndDate:     time.Now().AddDate(0, 0, 7),
			TotalDays:   5.0,
			Status:      "pending",
			Reason:      "Overlapping vacation",
		}

		err = svc.RequestLeave(context.Background(), overlapRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "overlapping leave request exists")

		// Test creating request with invalid dates
		invalidDatesRequest := &models.LeaveRequest{
			EmployeeID:  employee.ID,
			LeaveTypeID: leaveType.ID,
			StartDate:   time.Now().AddDate(0, 0, 5),
			EndDate:     time.Now().AddDate(0, 0, 1), // End date before start date
			TotalDays:   5.0,
			Status:      "pending",
			Reason:      "Invalid dates",
		}

		err = svc.RequestLeave(context.Background(), invalidDatesRequest)
		assert.Error(t, err)
	})

	t.Run("RecordAttendance", func(t *testing.T) {
		// Create test employee
		dept := testutil.GenerateDepartment(0)
		err := svc.CreateDepartment(context.Background(), dept)
		require.NoError(t, err)

		position := &models.JobPosition{
			Code:         "JP001",
			Title:        "Software Engineer",
			DepartmentID: dept.ID,
			GradeLevel:   "L4",
			SalaryMin:    50000.0,
			SalaryMax:    100000.0,
			Requirements: "5 years experience",
			IsActive:     true,
		}
		err = svc.CreateJobPosition(context.Background(), position)
		require.NoError(t, err)

		employee := testutil.GenerateEmployee(0)
		employee.DepartmentID = dept.ID
		employee.JobPositionID = position.ID
		err = svc.CreateEmployee(context.Background(), employee)
		require.NoError(t, err)

		// Test recording valid attendance
		now := time.Now()
		nowPtr := now
		checkOutTime := now.Add(9 * time.Hour)
		record := &models.AttendanceRecord{
			EmployeeID: employee.ID,
			Date:       now,
			CheckIn:    &nowPtr,
			CheckOut:   &checkOutTime,
			Status:     "present",
			Notes:      "Regular workday",
		}

		err = svc.RecordAttendance(context.Background(), record)
		assert.NoError(t, err)
		assert.NotZero(t, record.ID)

		// Test recording duplicate attendance
		dupRecord := &models.AttendanceRecord{
			EmployeeID: employee.ID,
			Date:       now,
			CheckIn:    &nowPtr,
			CheckOut:   &checkOutTime,
			Status:     "present",
			Notes:      "Duplicate attendance",
		}

		err = svc.RecordAttendance(context.Background(), dupRecord)
		assert.Error(t, err)

		// Test recording attendance with invalid checkout time
		invalidCheckout := now.Add(-1 * time.Hour)
		invalidRecord := &models.AttendanceRecord{
			EmployeeID: employee.ID,
			Date:       now,
			CheckIn:    &nowPtr,
			CheckOut:   &invalidCheckout,
			Status:     "present",
			Notes:      "Invalid checkout time",
		}

		err = svc.RecordAttendance(context.Background(), invalidRecord)
		assert.Error(t, err)
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		// Create test department
		dept := testutil.GenerateDepartment(0)
		err := svc.CreateDepartment(context.Background(), dept)
		require.NoError(t, err)

		// Try to create an employee with invalid data to trigger rollback
		employee := testutil.GenerateEmployee(0)
		employee.DepartmentID = dept.ID
		employee.JobPositionID = 999999 // Invalid position ID
		err = svc.CreateEmployee(context.Background(), employee)
		assert.Error(t, err)

		// Verify no salary components were created
		var count int
		err = testDB.DB.DB().QueryRowContext(context.Background(),
			"SELECT COUNT(*) FROM employee_salary WHERE employee_id = ?",
			employee.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("ConcurrentLeaveRequests", func(t *testing.T) {
		// Create test employee and leave type
		dept := testutil.GenerateDepartment(0)
		err := svc.CreateDepartment(context.Background(), dept)
		require.NoError(t, err)

		position := &models.JobPosition{
			Code:         "JP001",
			Title:        "Software Engineer",
			DepartmentID: dept.ID,
			GradeLevel:   "L4",
			SalaryMin:    50000.0,
			SalaryMax:    100000.0,
			Requirements: "5 years experience",
			IsActive:     true,
		}
		err = svc.CreateJobPosition(context.Background(), position)
		require.NoError(t, err)

		employee := testutil.GenerateEmployee(0)
		employee.DepartmentID = dept.ID
		employee.JobPositionID = position.ID
		err = svc.CreateEmployee(context.Background(), employee)
		require.NoError(t, err)

		leaveType := &models.LeaveType{
			Code:            "AL",
			Name:            "Annual Leave",
			Description:     "Regular annual leave",
			Paid:            true,
			AnnualAllowance: 20.0,
			IsActive:        true,
		}
		err = svc.CreateLeaveType(context.Background(), leaveType)
		require.NoError(t, err)

		// Submit multiple leave requests concurrently
		done := make(chan bool)
		startDate := time.Now().AddDate(0, 1, 0) // Start next month
		for i := 0; i < 5; i++ {
			go func(i int) {
				request := &models.LeaveRequest{
					EmployeeID:  employee.ID,
					LeaveTypeID: leaveType.ID,
					StartDate:   startDate.AddDate(0, 0, i*7),   // Different weeks
					EndDate:     startDate.AddDate(0, 0, i*7+2), // 3 days each
					TotalDays:   3.0,
					Status:      "pending",
					Reason:      fmt.Sprintf("Vacation %d", i),
				}
				err := svc.RequestLeave(context.Background(), request)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-done
		}

		// Verify total leave days don't exceed allowance
		var totalDays float64
		rows, err := testDB.DB.DB().QueryContext(context.Background(),
			`SELECT COALESCE(SUM(total_days), 0)
			FROM leave_requests
			WHERE employee_id = ? AND status != 'rejected'
			AND YEAR(start_date) = YEAR(CURRENT_DATE())`,
			employee.ID)
		require.NoError(t, err)
		defer rows.Close()

		require.True(t, rows.Next())
		err = rows.Scan(&totalDays)
		require.NoError(t, err)
		assert.LessOrEqual(t, totalDays, leaveType.AnnualAllowance)
	})
}
