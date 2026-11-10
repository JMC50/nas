package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// TokensValidAfter returns the time (Unix milliseconds) before which the user's
// login tokens are rejected. 0 means no restriction. found is false when no
// user has this userId.
func TokensValidAfter(conn *sql.DB, userID string) (validAfter int64, found bool, err error) {
	var value sql.NullInt64
	err = conn.QueryRow("SELECT tokens_valid_after FROM users WHERE userId = ?", userID).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("scan tokens_valid_after: %w", err)
	}
	return value.Int64, true, nil
}

func SetTokensValidAfter(conn *sql.DB, primaryKey, validAfter int64) error {
	_, err := conn.Exec("UPDATE users SET tokens_valid_after = ? WHERE id = ?", validAfter, primaryKey)
	if err != nil {
		return fmt.Errorf("update tokens_valid_after: %w", err)
	}
	return nil
}
