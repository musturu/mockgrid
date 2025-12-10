// Package sqlite provides a SQLite-based message store implementation.
package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mustur/mockgrid/app/api/store"
)

// Store persists messages in a SQLite database.
type Store struct {
	db *sql.DB
}

// New creates a new SQLite store at the given path.
func New(dbPath string) (s *Store, err error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	s = &Store{db: db}
	// Ensure db.Close() is deferred in case of an error
	defer func() {
		if err != nil {
			err = db.Close()
		}
	}()

	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	// err will be either be populated from db.close or be nil here
	return s, err
}

// Save inserts or updates a message in the database.
func (s *Store) Save(msg *store.Message) error {
	if msg.MsgID == "" {
		return fmt.Errorf("message ID is required")
	}

	query := `
INSERT INTO messages (
msg_id, from_email, to_email, subject, html_body, text_body,
status, smtp_response, reason, timestamp, last_event_time,
opens_count, clicks_count
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(msg_id) DO UPDATE SET
status = excluded.status,
smtp_response = excluded.smtp_response,
reason = excluded.reason,
last_event_time = excluded.last_event_time,
opens_count = excluded.opens_count,
clicks_count = excluded.clicks_count
`

	_, err := s.db.Exec(query,
		msg.MsgID, msg.FromEmail, msg.ToEmail, msg.Subject,
		msg.HTMLBody, msg.TextBody, msg.Status, msg.SMTPResponse,
		msg.Reason, msg.Timestamp, msg.LastEventTime,
		msg.OpensCount, msg.ClicksCount,
	)
	if err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	return nil
}

// Get retrieves messages based on query parameters.
func (s *Store) Get(query store.GetQuery) ([]*store.Message, error) {
	if query.ID != "" {
		return s.getByID(query.ID)
	}
	return s.list(query)
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	query := `
CREATE TABLE IF NOT EXISTS messages (
msg_id TEXT PRIMARY KEY,
from_email TEXT NOT NULL,
to_email TEXT NOT NULL,
subject TEXT,
html_body TEXT,
text_body TEXT,
status TEXT NOT NULL,
smtp_response TEXT,
reason TEXT,
timestamp INTEGER NOT NULL,
last_event_time INTEGER,
opens_count INTEGER DEFAULT 0,
clicks_count INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
`
	_, err := s.db.Exec(query)
	return err
}

func (s *Store) getByID(id string) ([]*store.Message, error) {
	query := `
SELECT msg_id, from_email, to_email, subject, html_body, text_body,
       status, smtp_response, reason, timestamp, last_event_time,
       opens_count, clicks_count
FROM messages WHERE msg_id = ?
`

	row := s.db.QueryRow(query, id)
	msg, err := s.scanMessage(row)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query message: %w", err)
	}

	return []*store.Message{msg}, nil
}

func (s *Store) list(query store.GetQuery) ([]*store.Message, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 100
	}

	var rows *sql.Rows
	var err error

	baseQuery := `
SELECT msg_id, from_email, to_email, subject, html_body, text_body,
       status, smtp_response, reason, timestamp, last_event_time,
       opens_count, clicks_count
FROM messages
`

	if query.Status != "" {
		rows, err = s.db.Query(
			baseQuery+" WHERE status = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?",
			query.Status, limit, query.Offset,
		)
	} else {
		rows, err = s.db.Query(
			baseQuery+" ORDER BY timestamp DESC LIMIT ? OFFSET ?",
			limit, query.Offset,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	var messages []*store.Message
	for rows.Next() {
		msg, err := s.scanMessageRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	return messages, nil
}

func (s *Store) scanMessage(row *sql.Row) (*store.Message, error) {
	var msg store.Message
	err := row.Scan(
		&msg.MsgID, &msg.FromEmail, &msg.ToEmail, &msg.Subject,
		&msg.HTMLBody, &msg.TextBody, &msg.Status, &msg.SMTPResponse,
		&msg.Reason, &msg.Timestamp, &msg.LastEventTime,
		&msg.OpensCount, &msg.ClicksCount,
	)
	return &msg, err
}

func (s *Store) scanMessageRows(rows *sql.Rows) (*store.Message, error) {
	var msg store.Message
	err := rows.Scan(
		&msg.MsgID, &msg.FromEmail, &msg.ToEmail, &msg.Subject,
		&msg.HTMLBody, &msg.TextBody, &msg.Status, &msg.SMTPResponse,
		&msg.Reason, &msg.Timestamp, &msg.LastEventTime,
		&msg.OpensCount, &msg.ClicksCount,
	)
	return &msg, err
}
