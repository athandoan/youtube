package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/athandoan/youtube/metadata-service/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteRepo struct {
	DB *sql.DB
}

func NewSQLiteRepository(dbPath string) (domain.VideoRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Init Schema
	schema := `
	CREATE TABLE IF NOT EXISTS videos (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		bucket_name TEXT NOT NULL,
		object_key TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS videos_fts USING fts5(id UNINDEXED, title, description);

	CREATE TRIGGER IF NOT EXISTS videos_ai AFTER INSERT ON videos BEGIN
		INSERT INTO videos_fts(id, title, description) VALUES (new.id, new.title, new.description);
	END;

	CREATE TRIGGER IF NOT EXISTS videos_ad AFTER DELETE ON videos BEGIN
		DELETE FROM videos_fts WHERE id = old.id;
	END;

	CREATE TRIGGER IF NOT EXISTS videos_au AFTER UPDATE ON videos BEGIN
		UPDATE videos_fts SET title = new.title, description = new.description WHERE id = new.id;
	END;

	INSERT INTO videos_fts(id, title, description) 
	SELECT id, title, description FROM videos 
	WHERE id NOT IN (SELECT id FROM videos_fts);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &sqliteRepo{DB: db}, nil
}

func (r *sqliteRepo) List(ctx context.Context, query string) ([]*domain.Video, error) {
	sqlQuery := "SELECT id, title, status, created_at, bucket_name, object_key FROM videos WHERE status = 'ready'"
	var rows *sql.Rows
	var err error

	if query != "" {
		sqlQuery = `
			SELECT v.id, v.title, v.status, v.created_at, v.bucket_name, v.object_key 
			FROM videos v 
			JOIN videos_fts f ON v.id = f.id 
			WHERE v.status = 'ready' AND videos_fts MATCH ? 
			ORDER BY rank`
		rows, err = r.DB.QueryContext(ctx, sqlQuery, query)
	} else {
		rows, err = r.DB.QueryContext(ctx, sqlQuery)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*domain.Video
	for rows.Next() {
		var v domain.Video
		if err := rows.Scan(&v.ID, &v.Title, &v.Status, &v.CreatedAt, &v.BucketName, &v.ObjectKey); err != nil {
			log.Println("Scan error:", err)
			continue
		}
		videos = append(videos, &v)
	}
	return videos, nil
}

func (r *sqliteRepo) Create(ctx context.Context, v *domain.Video) error {
	_, err := r.DB.ExecContext(ctx, "INSERT INTO videos (id, title, bucket_name, object_key, status) VALUES (?, ?, ?, ?, 'pending')",
		v.ID, v.Title, v.BucketName, v.ObjectKey)
	return err
}

func (r *sqliteRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	res, err := r.DB.ExecContext(ctx, "UPDATE videos SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("video with id %s not found", id)
	}
	return nil
}

func (r *sqliteRepo) Get(ctx context.Context, id string) (*domain.Video, error) {
	var v domain.Video
	err := r.DB.QueryRowContext(ctx, "SELECT id, title, status, created_at, bucket_name, object_key FROM videos WHERE id = ?", id).
		Scan(&v.ID, &v.Title, &v.Status, &v.CreatedAt, &v.BucketName, &v.ObjectKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("video not found")
		}
		return nil, err
	}
	return &v, nil
}
