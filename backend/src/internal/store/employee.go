package store

import (
	"database/sql"
	"fmt"
	"welltaxpro/src/internal/types"

	"github.com/google/logger"
	"github.com/google/uuid"
)

// GetEmployeeByFirebaseUID retrieves an employee by their Firebase UID
func (s *Store) GetEmployeeByFirebaseUID(firebaseUID string) (*types.Employee, error) {
	query := `
		SELECT id, firebase_uid, email, first_name, last_name, role, is_active, created_at, updated_at
		FROM employees
		WHERE firebase_uid = $1 AND is_active = true
	`

	row := s.DB.QueryRow(query, firebaseUID)

	employee := &types.Employee{}
	err := row.Scan(
		&employee.ID,
		&employee.FirebaseUID,
		&employee.Email,
		&employee.FirstName,
		&employee.LastName,
		&employee.Role,
		&employee.IsActive,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("employee not found for firebase UID: %s", firebaseUID)
	}
	if err != nil {
		logger.Errorf("Failed to get employee by firebase UID: %v", err)
		return nil, err
	}

	return employee, nil
}

// GetEmployeeByID retrieves an employee by their ID
func (s *Store) GetEmployeeByID(employeeID uuid.UUID) (*types.Employee, error) {
	query := `
		SELECT id, firebase_uid, email, first_name, last_name, role, is_active, created_at, updated_at
		FROM employees
		WHERE id = $1
	`

	row := s.DB.QueryRow(query, employeeID)

	employee := &types.Employee{}
	err := row.Scan(
		&employee.ID,
		&employee.FirebaseUID,
		&employee.Email,
		&employee.FirstName,
		&employee.LastName,
		&employee.Role,
		&employee.IsActive,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("employee not found with ID: %s", employeeID)
	}
	if err != nil {
		logger.Errorf("Failed to get employee by ID: %v", err)
		return nil, err
	}

	return employee, nil
}

// CreateEmployee creates a new employee record
func (s *Store) CreateEmployee(firebaseUID, email string, firstName, lastName *string, role string) (*types.Employee, error) {
	query := `
		INSERT INTO employees (firebase_uid, email, first_name, last_name, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, firebase_uid, email, first_name, last_name, role, is_active, created_at, updated_at
	`

	employee := &types.Employee{}
	err := s.DB.QueryRow(query, firebaseUID, email, firstName, lastName, role).Scan(
		&employee.ID,
		&employee.FirebaseUID,
		&employee.Email,
		&employee.FirstName,
		&employee.LastName,
		&employee.Role,
		&employee.IsActive,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err != nil {
		logger.Errorf("Failed to create employee: %v", err)
		return nil, err
	}

	logger.Infof("Created employee: %s (%s)", employee.Email, employee.ID)
	return employee, nil
}

// UpdateEmployee updates an employee's information
func (s *Store) UpdateEmployee(employeeID uuid.UUID, firstName, lastName *string, role string) (*types.Employee, error) {
	query := `
		UPDATE employees
		SET first_name = $1, last_name = $2, role = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING id, firebase_uid, email, first_name, last_name, role, is_active, created_at, updated_at
	`

	employee := &types.Employee{}
	err := s.DB.QueryRow(query, firstName, lastName, role, employeeID).Scan(
		&employee.ID,
		&employee.FirebaseUID,
		&employee.Email,
		&employee.FirstName,
		&employee.LastName,
		&employee.Role,
		&employee.IsActive,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err != nil {
		logger.Errorf("Failed to update employee: %v", err)
		return nil, err
	}

	logger.Infof("Updated employee: %s", employee.ID)
	return employee, nil
}

// DeactivateEmployee marks an employee as inactive
func (s *Store) DeactivateEmployee(employeeID uuid.UUID) error {
	query := `
		UPDATE employees
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := s.DB.Exec(query, employeeID)
	if err != nil {
		logger.Errorf("Failed to deactivate employee: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("employee not found: %s", employeeID)
	}

	logger.Infof("Deactivated employee: %s", employeeID)
	return nil
}

// GetAllEmployees retrieves all employees
func (s *Store) GetAllEmployees(includeInactive bool) ([]*types.Employee, error) {
	query := `
		SELECT id, firebase_uid, email, first_name, last_name, role, is_active, created_at, updated_at
		FROM employees
	`

	if !includeInactive {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.DB.Query(query)
	if err != nil {
		logger.Errorf("Failed to get employees: %v", err)
		return nil, err
	}
	defer rows.Close()

	var employees []*types.Employee
	for rows.Next() {
		employee := &types.Employee{}
		err := rows.Scan(
			&employee.ID,
			&employee.FirebaseUID,
			&employee.Email,
			&employee.FirstName,
			&employee.LastName,
			&employee.Role,
			&employee.IsActive,
			&employee.CreatedAt,
			&employee.UpdatedAt,
		)
		if err != nil {
			logger.Errorf("Failed to scan employee: %v", err)
			return nil, err
		}
		employees = append(employees, employee)
	}

	return employees, rows.Err()
}
