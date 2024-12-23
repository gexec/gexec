package store

import (
	"errors"
)

var (
	// ErrWrongCredentials is returned when credentials are wrong.
	ErrWrongCredentials = errors.New("wrong credentials provided")

	// ErrAlreadyAssigned defines the error if relation is already assigned.
	ErrAlreadyAssigned = errors.New("user pack already exists")

	// ErrNotAssigned defines the error if relation is not assigned.
	ErrNotAssigned = errors.New("user pack is not defined")

	// ErrTeamNotFound is returned when a user was not found.
	ErrTeamNotFound = errors.New("team not found")

	// ErrUserNotFound is returned when a user was not found.
	ErrUserNotFound = errors.New("user not found")

	// ErrProjectNotFound is returned when a user was not found.
	ErrProjectNotFound = errors.New("project not found")
)
