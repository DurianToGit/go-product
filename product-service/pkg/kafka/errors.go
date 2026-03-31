package kafka

import (
	"errors"
	"fmt"
)

type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("retryable: %v", e.Err)
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return fmt.Sprintf("non-retryable: %v", e.Err)
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

func NewRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &RetryableError{Err: err}
}

func NewNonRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &NonRetryableError{Err: err}
}

func IsRetryable(err error) bool {
	var target *RetryableError
	return errors.As(err, &target)
}

func IsNonRetryable(err error) bool {
	var target *NonRetryableError
	return errors.As(err, &target)
}
