package domain

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidInput       = errors.New("invalid input")
	ErrTeamNotFound       = errors.New("team not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyInTeam  = errors.New("user already in team")
	ErrForbidden          = errors.New("insufficient permissions")
	ErrCannotInviteOwner  = errors.New("cannot invite user as owner")
	ErrInvalidRole        = errors.New("invalid role")
)
