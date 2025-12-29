package storage

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"

	"tokenlaunch/internal/domain"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(dsn string) (*Postgres, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) Save(ctx context.Context, msg domain.Message) error {
	query := `
		INSERT INTO messages (id, external_id, author, username, content, source, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`

	_, err := p.db.ExecContext(ctx, query,
		msg.ID,
		msg.ExternalID,
		msg.Author,
		msg.Username,
		msg.Content,
		msg.Source,
		msg.CreatedAt,
	)

	return err
}

func (p *Postgres) FindByID(ctx context.Context, id string) (*domain.Message, error) {
	query := `
		SELECT id, external_id, author, username, content, source, created_at
		FROM messages WHERE id = $1
	`

	var msg domain.Message
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID,
		&msg.ExternalID,
		&msg.Author,
		&msg.Username,
		&msg.Content,
		&msg.Source,
		&msg.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (p *Postgres) FindAll(ctx context.Context, limit, offset int) ([]domain.Message, error) {
	query := `
		SELECT id, external_id, author, username, content, source, created_at
		FROM messages ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := p.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.ExternalID,
			&msg.Author,
			&msg.Username,
			&msg.Content,
			&msg.Source,
			&msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

func (p *Postgres) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM messages WHERE id = $1)`

	var exists bool
	err := p.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}
