package db

import (
	"database/sql"
	"fmt"
)

const intentAdmin = "ADMIN"

// HasIntent returns true if the user has the specified intent OR the ADMIN intent.
// ADMIN is a super-permission that grants everything.
func HasIntent(conn *sql.DB, userID, intent string) (bool, error) {
	row := conn.QueryRow("SELECT id FROM users WHERE userId = ?", userID)
	var primaryKey int64
	if err := row.Scan(&primaryKey); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("find user: %w", err)
	}

	var hit int
	err := conn.QueryRow(`
		SELECT 1 FROM user_intents
		WHERE user_id = ? AND intent IN (?, ?) LIMIT 1`,
		primaryKey, intent, intentAdmin).Scan(&hit)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check intent: %w", err)
	}
	return true, nil
}

// ToggleIntent flips an intent for a user: deletes if present, inserts if not.
// Mirrors Node `editIntent` semantics.
func ToggleIntent(conn *sql.DB, userID, intent string) error {
	row := conn.QueryRow("SELECT id FROM users WHERE userId = ?", userID)
	var primaryKey int64
	if err := row.Scan(&primaryKey); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("find user: %w", err)
	}

	var existing int
	err := conn.QueryRow(`
		SELECT 1 FROM user_intents WHERE user_id = ? AND intent = ?`,
		primaryKey, intent).Scan(&existing)
	switch err {
	case sql.ErrNoRows:
		_, err := conn.Exec(`
			INSERT INTO user_intents (user_id, intent) VALUES (?, ?)`,
			primaryKey, intent)
		if err != nil {
			return fmt.Errorf("insert intent: %w", err)
		}
	case nil:
		_, err := conn.Exec(`
			DELETE FROM user_intents WHERE user_id = ? AND intent = ?`,
			primaryKey, intent)
		if err != nil {
			return fmt.Errorf("delete intent: %w", err)
		}
	default:
		return fmt.Errorf("check intent: %w", err)
	}
	return nil
}

func userIntents(conn *sql.DB, primaryKey int64) ([]string, error) {
	rows, err := conn.Query("SELECT intent FROM user_intents WHERE user_id = ?", primaryKey)
	if err != nil {
		return nil, fmt.Errorf("query intents: %w", err)
	}
	defer rows.Close()
	intents := []string{}
	for rows.Next() {
		var intent string
		if err := rows.Scan(&intent); err != nil {
			return nil, fmt.Errorf("scan intent: %w", err)
		}
		intents = append(intents, intent)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return intents, nil
}

func allUserIntents(conn *sql.DB) (map[int64][]string, error) {
	rows, err := conn.Query("SELECT user_id, intent FROM user_intents")
	if err != nil {
		return nil, fmt.Errorf("query all intents: %w", err)
	}
	defer rows.Close()
	intentMap := map[int64][]string{}
	for rows.Next() {
		var userID int64
		var intent string
		if err := rows.Scan(&userID, &intent); err != nil {
			return nil, fmt.Errorf("scan intent row: %w", err)
		}
		intentMap[userID] = append(intentMap[userID], intent)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return intentMap, nil
}
