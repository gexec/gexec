package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/genexec/genexec/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

// Auth provides all database operations related to auth.
type Auth struct {
	client *Store
}

// ByID tries to authenticate a user based on identifier.
func (s *Auth) ByID(ctx context.Context, userID string) (*model.User, error) {
	record := &model.User{}

	if err := s.client.handle.NewSelect().
		Model(record).
		Relation("Auths").
		Relation("Teams").
		Relation("Teams.Team").
		Relation("Teams.Team.Projects").
		Relation("Projects").
		Where("id = ?", userID).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrUserNotFound
		}

		return record, err
	}

	return record, nil
}

// ByCreds tries to authenticate a user based on credentials.
func (s *Auth) ByCreds(ctx context.Context, username, password string) (*model.User, error) {
	record := &model.User{}

	if err := s.client.handle.NewSelect().
		Model(record).
		Relation("Auths").
		Relation("Teams").
		Relation("Teams.Team").
		Relation("Teams.Team.Projects").
		Relation("Projects").
		Where("username = ?", username).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrUserNotFound
		}

		return record, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(record.Hashword),
		[]byte(password),
	); err != nil {
		return nil, ErrWrongCredentials
	}

	return record, nil
}
