package errors

import "errors"

var (
	ErrDatabaseServer = errors.New("database server error")
	ErrDatabase       = errors.New("database error")
	ErrDatabaseExists = errors.New("database already exists")

	// Redis Cache
	ErrRedisServer = errors.New("redis server error")
)
