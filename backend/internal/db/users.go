package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// User mirrors the Node `users` row + denormalized `intents` slice (joined separately).
type User struct {
	ID         int64
	UserID     string
	Username   string
	GlobalName string
	KrName     string
	Password   string
	AuthType   string
	Intents    []string
}

func GetUser(conn *sql.DB, userID string) (*User, error) {
	row := conn.QueryRow(`
		SELECT id, userId, username, COALESCE(global_name, ''), COALESCE(krname, ''),
		       COALESCE(password, ''), COALESCE(auth_type, '')
		FROM users WHERE userId = ?`, userID)
	user := &User{}
	if err := row.Scan(&user.ID, &user.UserID, &user.Username, &user.GlobalName,
		&user.KrName, &user.Password, &user.AuthType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	intents, err := userIntents(conn, user.ID)
	if err != nil {
		return nil, err
	}
	user.Intents = intents
	return user, nil
}

func GetUserByPrimaryKey(conn *sql.DB, id int64) (*User, error) {
	row := conn.QueryRow(`
		SELECT id, userId, username, COALESCE(global_name, ''), COALESCE(krname, ''),
		       COALESCE(password, ''), COALESCE(auth_type, '')
		FROM users WHERE id = ?`, id)
	user := &User{}
	if err := row.Scan(&user.ID, &user.UserID, &user.Username, &user.GlobalName,
		&user.KrName, &user.Password, &user.AuthType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	intents, err := userIntents(conn, user.ID)
	if err != nil {
		return nil, err
	}
	user.Intents = intents
	return user, nil
}

func GetAllUsers(conn *sql.DB) ([]*User, error) {
	rows, err := conn.Query(`
		SELECT id, userId, username, COALESCE(global_name, ''), COALESCE(krname, ''),
		       COALESCE(password, ''), COALESCE(auth_type, '')
		FROM users`)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.UserID, &user.Username, &user.GlobalName,
			&user.KrName, &user.Password, &user.AuthType); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	intentMap, err := allUserIntents(conn)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		user.Intents = intentMap[user.ID]
		if user.Intents == nil {
			user.Intents = []string{}
		}
	}
	return users, nil
}

// SaveOAuthUser inserts an OAuth user (Discord/Google) or updates existing.
// Returns the user's primary key.
func SaveOAuthUser(conn *sql.DB, userID, username, globalName, krName string) (int64, error) {
	row := conn.QueryRow("SELECT id FROM users WHERE userId = ?", userID)
	var existingID int64
	err := row.Scan(&existingID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		result, err := conn.Exec(`
			INSERT INTO users (userId, username, global_name, krname)
			VALUES (?, ?, ?, ?)`, userID, username, globalName, krName)
		if err != nil {
			return 0, fmt.Errorf("insert user: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("get insert id: %w", err)
		}
		return id, nil
	case err != nil:
		return 0, fmt.Errorf("check existing user: %w", err)
	default:
		_, err := conn.Exec(`
			UPDATE users SET username = ?, global_name = ?, krname = ?
			WHERE userId = ?`, username, globalName, krName, userID)
		if err != nil {
			return 0, fmt.Errorf("update user: %w", err)
		}
		return existingID, nil
	}
}

// SaveLocalUser inserts a local-auth user with bcrypt password hash.
// Fails if userId already exists (caller should check first).
func SaveLocalUser(conn *sql.DB, userID, username, passwordHash, krName string) (int64, error) {
	result, err := conn.Exec(`
		INSERT INTO users (userId, username, password, krname, global_name, auth_type)
		VALUES (?, ?, ?, ?, ?, 'local')`,
		userID, username, passwordHash, krName, username)
	if err != nil {
		return 0, fmt.Errorf("insert local user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get insert id: %w", err)
	}
	return id, nil
}

func UpdatePassword(conn *sql.DB, userID, newHash string) error {
	_, err := conn.Exec("UPDATE users SET password = ? WHERE userId = ?", newHash, userID)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

// ErrUserNotFound is returned by DeleteUser when no row matches the supplied userID.
var ErrUserNotFound = errors.New("user not found")

// DeleteUser hard-deletes the user row plus their activity log entries.
// FK cascades drop user_intents and user_identities. The log table declares
// `user_id NOT NULL` with `ON DELETE SET NULL`, which is a latent schema bug
// (SET NULL fails the NOT NULL constraint), so we delete log rows explicitly
// inside the same transaction rather than relying on the broken cascade.
func DeleteUser(conn *sql.DB, userID string) error {
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var primaryKey int64
	if err := tx.QueryRow("SELECT id FROM users WHERE userId = ?", userID).Scan(&primaryKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("lookup user: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM log WHERE user_id = ?", primaryKey); err != nil {
		return fmt.Errorf("delete user logs: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM users WHERE id = ?", primaryKey); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete: %w", err)
	}
	return nil
}
