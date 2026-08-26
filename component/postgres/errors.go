package postgres

import (
	"errors"
	"fmt"

	types "go.wasmcloud.dev/component/imports/wasmcloud_postgres_0_2_0_types"
)

// Sentinel errors for the argument-free named cases of
// wasmcloud:postgres/types.error. Test for them with errors.Is.
var (
	ErrUnknownPreparedStatement = errors.New("postgres: unknown prepared statement")
	ErrAccessDenied             = errors.New("postgres: access denied")
	ErrTimeout                  = errors.New("postgres: timeout")
)

// DBError is a structured error reported by the database itself, carried by
// errors returned from this package when the database rejected or failed a
// statement. Retrieve it with errors.As and branch on the machine-readable
// Code (the five-character SQLSTATE, e.g. "23505" for a unique-constraint
// violation) rather than parsing Message.
type DBError struct {
	// Code is the five-character SQLSTATE code.
	Code string
	// Severity is the severity reported by the database (ERROR, FATAL, ...).
	Severity string
	// Message is the primary human-readable error message.
	Message string
	// Detail optionally carries more detail about the problem.
	Detail string
	// Extras holds any additional (field, value) pairs the database provided.
	Extras [][2]string
}

func (e *DBError) Error() string {
	return fmt.Sprintf("postgres: query failed: %s %s: %s", e.Severity, e.Code, e.Message)
}

func convertError(e types.Error) error {
	switch e.Tag() {
	case types.ErrorConnectionFailed:
		return fmt.Errorf("postgres: connection failed: %s", e.ConnectionFailed())
	case types.ErrorInvalidParams:
		return fmt.Errorf("postgres: invalid params: %s", e.InvalidParams())
	case types.ErrorInvalidQuery:
		return fmt.Errorf("postgres: invalid query: %s", e.InvalidQuery())
	case types.ErrorQueryFailed:
		db := e.QueryFailed()
		out := &DBError{
			Code:     db.Code,
			Severity: db.Severity,
			Message:  db.Message,
			Detail:   db.Detail.SomeOr(""),
		}
		for _, extra := range db.Extras {
			out.Extras = append(out.Extras, [2]string{extra.F0, extra.F1})
		}
		return out
	case types.ErrorValueConversionFailed:
		return fmt.Errorf("postgres: value conversion failed: %s", e.ValueConversionFailed())
	case types.ErrorUnknownPreparedStatement:
		return ErrUnknownPreparedStatement
	case types.ErrorAccessDenied:
		return ErrAccessDenied
	case types.ErrorTimeout:
		return ErrTimeout
	case types.ErrorOther:
		return fmt.Errorf("postgres: %s", e.Other())
	default:
		// The variant is non-exhaustive; future hosts may add cases.
		return fmt.Errorf("postgres: unknown error (tag %d)", e.Tag())
	}
}
