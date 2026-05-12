package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type LogEntry struct {
	Activity    string `json:"activity"`
	Description string `json:"description"`
	Time        int64  `json:"time"`
	Loc         string `json:"loc"`
	UserID      string `json:"userId,omitempty"`
	Username    string `json:"username,omitempty"`
	KrName      string `json:"krname,omitempty"`
}

func InsertLog(conn *sql.DB, userID, activity, description, loc string, timeMs int64) error {
	row := conn.QueryRow("SELECT id FROM users WHERE userId = ?", userID)
	var primaryKey int64
	if err := row.Scan(&primaryKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user %s not found for logging", userID)
		}
		return fmt.Errorf("find user: %w", err)
	}
	_, err := conn.Exec(`
		INSERT INTO log (activity, description, user_id, time, loc)
		VALUES (?, ?, ?, ?, ?)`,
		activity, description, primaryKey, timeMs, loc)
	if err != nil {
		return fmt.Errorf("insert log: %w", err)
	}
	return nil
}

func GetActivityLogs(conn *sql.DB) ([]LogEntry, error) {
	rows, err := conn.Query(`
		SELECT log.activity, log.description, log.time, log.loc,
		       COALESCE(users.userId, '') as userId,
		       COALESCE(users.username, '') as username,
		       COALESCE(users.krname, '') as krname
		FROM log
		LEFT JOIN users ON log.user_id = users.id
		ORDER BY log.time DESC`)
	if err != nil {
		return nil, fmt.Errorf("query logs: %w", err)
	}
	defer rows.Close()
	logs := []LogEntry{}
	for rows.Next() {
		var entry LogEntry
		if err := rows.Scan(&entry.Activity, &entry.Description, &entry.Time, &entry.Loc,
			&entry.UserID, &entry.Username, &entry.KrName); err != nil {
			return nil, fmt.Errorf("scan log: %w", err)
		}
		logs = append(logs, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}
