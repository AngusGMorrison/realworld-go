// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.21.0

package sqlc

import (
	"database/sql"
	"time"
)

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	Bio          sql.NullString
	ImageUrl     sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
