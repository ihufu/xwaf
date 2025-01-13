package repository

import "errors"

// ErrNotFound 表示资源未找到
var ErrNotFound = errors.New("resource not found")

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// ... existing code ...
