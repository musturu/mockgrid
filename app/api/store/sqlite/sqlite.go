// Package sqlite provides a SQLite-based message store implementation.
package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mustur/mockgrid/app/api/store"
)

// Store persists messages in a SQLite database.
type Store struct {
	path string
	db   *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return &Store{path: path, db: db}, nil
}

// Connect creates a new SQLite store at the given path.
func (s *Store) Connect() error {

	if err := s.migrate(); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
}

// Save inserts or updates a message in the database.
func (s *Store) SaveMSG(msg *store.Message) error {
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
func (s *Store) GetMSG(query store.GetQuery) ([]*store.Message, error) {
	if query.ID != "" {
		return s.getMSGByID(query.ID)
	}
	return s.listMSG(query)
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
CREATE TABLE IF NOT EXISTS webhooks (
	id TEXT PRIMARY KEY,
	url TEXT NOT NULL,
	events TEXT NOT NULL,
	enabled BOOLEAN NOT NULL DEFAULT 1,
	secret TEXT,
	created_at INTEGER,
	updated_at INTEGER
);
`
	_, err := s.db.Exec(query)
	return err
}

// WebhookStore implementation
func (s *Store) Create(hook *store.WebhookConfig) error {
	eventsJSON, err := json.Marshal(hook.Events)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO webhooks (id, url, events, enabled, secret, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		hook.ID, hook.URL, string(eventsJSON), hook.Enabled, hook.Secret, hook.CreatedAt, hook.UpdatedAt)
	return err
}

func (s *Store) GetWebhook(id string) (*store.WebhookConfig, error) {
	var cfg store.WebhookConfig
	var eventsJSON string
	err := s.db.QueryRow(`SELECT id, url, events, enabled, secret, created_at, updated_at FROM webhooks WHERE id = ?`, id).
		Scan(&cfg.ID, &cfg.URL, &eventsJSON, &cfg.Enabled, &cfg.Secret, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(eventsJSON), &cfg.Events); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Store) ListWebhooks() ([]*store.WebhookConfig, error) {
	rows, err := s.db.Query(`SELECT id, url, events, enabled, secret, created_at, updated_at FROM webhooks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*store.WebhookConfig
	for rows.Next() {
		var cfg store.WebhookConfig
		var eventsJSON string
		if err := rows.Scan(&cfg.ID, &cfg.URL, &eventsJSON, &cfg.Enabled, &cfg.Secret, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(eventsJSON), &cfg.Events); err != nil {
			return nil, err
		}
		res = append(res, &cfg)
	}
	return res, nil
}

func (s *Store) ListEnabledWebhooks() ([]*store.WebhookConfig, error) {
	rows, err := s.db.Query(`SELECT id, url, events, enabled, secret, created_at, updated_at FROM webhooks WHERE enabled = 1 ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*store.WebhookConfig
	for rows.Next() {
		var cfg store.WebhookConfig
		var eventsJSON string
		if err := rows.Scan(&cfg.ID, &cfg.URL, &eventsJSON, &cfg.Enabled, &cfg.Secret, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(eventsJSON), &cfg.Events); err != nil {
			return nil, err
		}
		res = append(res, &cfg)
	}
	return res, nil
}

func (s *Store) UpdateWebhook(hook *store.WebhookConfig) error {
	eventsJSON, err := json.Marshal(hook.Events)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE webhooks SET url = ?, events = ?, enabled = ?, secret = ?, updated_at = ? WHERE id = ?`,
		hook.URL, string(eventsJSON), hook.Enabled, hook.Secret, hook.UpdatedAt, hook.ID)
	return err
}

func (s *Store) DeleteWebhook(id string) error {
	_, err := s.db.Exec(`DELETE FROM webhooks WHERE id = ?`, id)
	return err
}

func (s *Store) getMSGByID(id string) ([]*store.Message, error) {
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

func (s *Store) listMSG(query store.GetQuery) ([]*store.Message, error) {
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
