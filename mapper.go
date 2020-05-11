package zergrepo

import (
	"errors"

	"github.com/lib/pq"
)

// ErrMapFunc converts certain errors into others.
// It is necessary for an opportunity not to lift errors of a database in the top layers.
type ErrMapFunc func(error) error

// Mapperer responsible for converting some types of errors into others.
type Mapperer interface {
	Map(err error) error
}

// Mapper is implements Mapperer.
type Mapper struct {
	converters []ErrMapFunc
}

// NewMapper create a new instance Mapper.
func NewMapper(converters ...ErrMapFunc) *Mapper {
	return &Mapper{
		converters: converters,
	}
}

// NewConvert returns the converter function.
func NewConvert(target error, variables ...error) ErrMapFunc {
	return func(err error) error {
		for i := range variables {
			if errors.Is(err, variables[i]) {
				return target
			}
		}
		return nil
	}
}

// PQConstraint returns a postgres oriented converter.
func PQConstraint(target error, constraint string) ErrMapFunc {
	return func(err error) error {
		pqErr, ok := err.(*pq.Error)
		if !ok {
			return nil
		}

		if pqErr.Constraint == constraint {
			return target
		}

		return nil
	}
}

// Map looking for one of all the functions that can convert an error.
// If it is not found, it will return the original error.
func (m *Mapper) Map(err error) error {
	for i := range m.converters {
		convertErr := m.converters[i](err)
		if convertErr != nil {
			return convertErr
		}
	}

	return err
}
