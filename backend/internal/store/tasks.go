package store

import (
	"database/sql"
	"fmt"

	"github.com/alkaid/goapiscanner/internal/model"
)

func (s *Store) InsertTask(t model.Task) error {
	auth := 0
	if t.Authorized {
		auth = 1
	}
	_, err := s.db.Exec(`INSERT INTO tasks(
		id,base_url,status,concurrency,timeout_ms,authorized,swagger_name,
		total,sent,hits,critical,high,medium,low,info,error,created_at,updated_at
	) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.BaseURL, t.Status, t.Concurrency, t.TimeoutMS, auth, t.SwaggerName,
		t.Total, t.Sent, t.Hits, t.Critical, t.High, t.Medium, t.Low, t.Info, t.Error, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (s *Store) UpdateTask(t model.Task) error {
	_, err := s.db.Exec(`UPDATE tasks SET
		status=?, total=?, sent=?, hits=?, critical=?, high=?, medium=?, low=?, info=?, error=?, updated_at=?, swagger_name=?
		WHERE id=?`,
		t.Status, t.Total, t.Sent, t.Hits, t.Critical, t.High, t.Medium, t.Low, t.Info, t.Error, t.UpdatedAt, t.SwaggerName, t.ID,
	)
	return err
}

func (s *Store) GetTask(id string) (model.Task, error) {
	row := s.db.QueryRow(`SELECT id,base_url,status,concurrency,timeout_ms,authorized,swagger_name,
		total,sent,hits,critical,high,medium,low,info,error,created_at,updated_at
		FROM tasks WHERE id=?`, id)
	t, err := scanTask(row)
	if err == sql.ErrNoRows {
		return model.Task{}, fmt.Errorf("task not found")
	}
	return t, err
}

func (s *Store) ListTasks() ([]model.Task, error) {
	rows, err := s.db.Query(`SELECT id,base_url,status,concurrency,timeout_ms,authorized,swagger_name,
		total,sent,hits,critical,high,medium,low,info,error,created_at,updated_at
		FROM tasks ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(sc rowScanner) (model.Task, error) {
	var t model.Task
	var auth int
	err := sc.Scan(&t.ID, &t.BaseURL, &t.Status, &t.Concurrency, &t.TimeoutMS, &auth, &t.SwaggerName,
		&t.Total, &t.Sent, &t.Hits, &t.Critical, &t.High, &t.Medium, &t.Low, &t.Info, &t.Error, &t.CreatedAt, &t.UpdatedAt)
	t.Authorized = auth == 1
	return t, err
}
