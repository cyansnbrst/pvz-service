package db

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicatePVZ      = errors.New("duplicate pvz")
	ErrReceptionConflict = errors.New("either pvz not found or previous reception still open")
	ErrNoOpenReception   = errors.New("no opened reception for the pvz was found")
	ErrNoProducts        = errors.New("no products in the reception")
)
