package errors

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidPrice      = errors.New("price must be non-negative")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidQuantity   = errors.New("quantity must be positive")
	ErrProductNotFound   = errors.New("product not found")
)

type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error { return e.Err }

func New(code, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Err: cause}
}

func Is(err, target error) bool { return errors.Is(err, target) }

func As(err error, target interface{}) bool { return errors.As(err, target) }
