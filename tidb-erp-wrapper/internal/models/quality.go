package models

import "time"

type QualityParameter struct {
	ID              int64     `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	MeasurementUnit string    `json:"measurement_unit"`
	MinValue        *float64  `json:"min_value,omitempty"`
	MaxValue        *float64  `json:"max_value,omitempty"`
	TargetValue     *float64  `json:"target_value,omitempty"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type InspectionPoint struct {
	ID           int64     `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ProcessType  string    `json:"process_type"`
	DepartmentID *int64    `json:"department_id,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type InspectionPlan struct {
	ID            int64      `json:"id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	ProductID     *int64     `json:"product_id,omitempty"`
	Version       string     `json:"version"`
	Status        string     `json:"status"`
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type InspectionPlanParameter struct {
	ID                int64     `json:"id"`
	InspectionPlanID  int64     `json:"inspection_plan_id"`
	InspectionPointID int64     `json:"inspection_point_id"`
	ParameterID       int64     `json:"parameter_id"`
	SamplingMethod    string    `json:"sampling_method"`
	SampleSize        *int      `json:"sample_size,omitempty"`
	Mandatory         bool      `json:"mandatory"`
	SequenceNo        int       `json:"sequence_no"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type QualityInspection struct {
	ID               int64     `json:"id"`
	InspectionNumber string    `json:"inspection_number"`
	InspectionPlanID int64     `json:"inspection_plan_id"`
	ReferenceType    string    `json:"reference_type"`
	ReferenceID      int64     `json:"reference_id"`
	InspectorID      int64     `json:"inspector_id"`
	InspectionDate   time.Time `json:"inspection_date"`
	Status           string    `json:"status"`
	Result           string    `json:"result"`
	Notes            string    `json:"notes"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type InspectionResult struct {
	ID            int64     `json:"id"`
	InspectionID  int64     `json:"inspection_id"`
	ParameterID   int64     `json:"parameter_id"`
	MeasuredValue *float64  `json:"measured_value,omitempty"`
	IsConforming  bool      `json:"is_conforming"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type QualityNotification struct {
	ID                 int64      `json:"id"`
	NotificationNumber string     `json:"notification_number"`
	Type               string     `json:"type"`
	Priority           string     `json:"priority"`
	ReferenceType      string     `json:"reference_type"`
	ReferenceID        int64      `json:"reference_id"`
	ReportedBy         int64      `json:"reported_by"`
	AssignedTo         *int64     `json:"assigned_to,omitempty"`
	Status             string     `json:"status"`
	Description        string     `json:"description"`
	RootCause          string     `json:"root_cause"`
	CorrectiveAction   string     `json:"corrective_action"`
	DueDate            *time.Time `json:"due_date,omitempty"`
	ClosedAt           *time.Time `json:"closed_at,omitempty"`
	ClosedBy           *int64     `json:"closed_by,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
