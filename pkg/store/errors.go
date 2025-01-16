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

	// ErrCredentialNotFound is returned when a credential was not found.
	ErrCredentialNotFound = errors.New("credential not found")

	// ErrRepositoryNotFound is returned when a repository was not found.
	ErrRepositoryNotFound = errors.New("repository not found")

	// ErrInventoryNotFound is returned when a inventory was not found.
	ErrInventoryNotFound = errors.New("inventory not found")

	// ErrEnvironmentNotFound is returned when a environment was not found.
	ErrEnvironmentNotFound = errors.New("environment not found")

	// ErrEnvironmentValueNotFound is returned when a environment value was not found.
	ErrEnvironmentValueNotFound = errors.New("environment value not found")

	// ErrEnvironmentSecretNotFound is returned when a environment secret was not found.
	ErrEnvironmentSecretNotFound = errors.New("environment secret not found")

	// ErrTemplateNotFound is returned when a template was not found.
	ErrTemplateNotFound = errors.New("template not found")

	// ErrTemplateSurveyNotFound is returned when a template survey was not found.
	ErrTemplateSurveyNotFound = errors.New("template survey not found")

	// ErrTemplateValueNotFound is returned when a template value was not found.
	ErrTemplateValueNotFound = errors.New("template value not found")

	// ErrTemplateVaultNotFound is returned when a template vault was not found.
	ErrTemplateVaultNotFound = errors.New("template vault not found")

	// ErrScheduleNotFound is returned when a schedule was not found.
	ErrScheduleNotFound = errors.New("schedule not found")

	// ErrExecutionNotFound is returned when a execution was not found.
	ErrExecutionNotFound = errors.New("execution not found")
)
