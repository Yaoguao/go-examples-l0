package storage

import "errors"

var (
	ErrURLNotFound = errors.New("order not found")
	ErrURLExists   = errors.New("order already exists")
)
