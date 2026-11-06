package device

import (
	"context"
	"database/sql"
	"errors"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("device not found")

const schema = `
CREATE TABLE IF NOT EXISTS devices (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	name       TEXT NOT NULL,
	brand      TEXT NOT NULL,
	created_at DATETIME NOT NULL
)`

type Store struct {
	db *sql.DB
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// sqlite handles one writer at a time
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Create(ctx context.Context, d *Device) error {
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO devices (name, brand, created_at) VALUES (?, ?, ?)",
		d.Name, d.Brand, d.CreatedAt)
	if err != nil {
		return err
	}
	d.ID, err = res.LastInsertId()
	return err
}

func (s *Store) Get(ctx context.Context, id int64) (Device, error) {
	var d Device
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, brand, created_at FROM devices WHERE id = ?", id).
		Scan(&d.ID, &d.Name, &d.Brand, &d.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return d, ErrNotFound
	}
	return d, err
}

func (s *Store) Update(ctx context.Context, d Device) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE devices SET name = ?, brand = ? WHERE id = ?",
		d.Name, d.Brand, d.ID)
	if err != nil {
		return err
	}
	return checkFound(res)
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM devices WHERE id = ?", id)
	if err != nil {
		return err
	}
	return checkFound(res)
}

func checkFound(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
