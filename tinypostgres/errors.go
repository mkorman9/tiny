package tinypostgres

import (
	"errors"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

const (
	// ErrUnknown is returned for all errors with unknown cause.
	ErrUnknown = iota

	// ErrUniqueViolation is returned for unique constraint violations.
	ErrUniqueViolation

	// ErrNotNullViolation is returned for non-null constraint violations.
	ErrNotNullViolation

	// ErrRecordNotFound is returned when query is expected to return a single record, but none are returned.
	ErrRecordNotFound

	// ErrInvalidText is returned when you try assigned invalid value to types such as uuid.
	ErrInvalidText
)

// Error represents a wrapped query error.
type Error struct {
	// Err is the wrapped error.
	Err error

	// Code is the error code.
	Code int

	// Constraint is the name of violated constraint.
	Constraint string

	// TableName is the name of the table.
	TableName string

	// ColumnName is the name of the column.
	ColumnName string

	// Message is the message returned by the DB.
	Message string
}

// TranslateError tries to cast given error instance to PgError and extract enough information to build Error instance.
func TranslateError(err error) *Error {
	result := &Error{
		Err:  err,
		Code: ErrUnknown,
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		result.Code = ErrRecordNotFound
		return result
	}

	if pgErr, ok := err.(*pgconn.PgError); ok {
		result.TableName = pgErr.TableName
		result.Constraint = pgErr.ConstraintName
		result.ColumnName = pgErr.ColumnName
		result.Message = pgErr.Message

		switch pgErr.Code {
		case "23502": // not_null_violation
			result.Code = ErrNotNullViolation
		case "23505": // unique_violation
			result.Code = ErrUniqueViolation
		case "22P02": // invalid_text_representation
			result.Code = ErrInvalidText
		}
	}

	return result
}
