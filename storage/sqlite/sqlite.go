package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"telegramBotSaver/storage"

	_ "github.com/mattn/go-sqlite3" //дравер для sqlite
)

type Storage struct {
	db *sql.DB
}

//path - путь до файла с бд
func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path) //уточняем с какой бд работаем

	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}

	//проверим получилось ли установить связь
	if err :=db.Ping(); err != nil{
		return nil, fmt.Errorf("can't open database: %w", err)
	}

	return &Storage {db: db,}, nil
}

func (s *Storage) Save(ctx context.Context, p *storage.Page) error {
	//создаем запрос
	q := `INSERT INTO pages (url, user_name) VALUES (?, ?)`

	//выполняем запрос
	_, err := s.db.ExecContext(ctx, q, p.URL, p.UserName)

	if err != nil {
		return fmt.Errorf("can't save page: %w", err)
	}

	return nil
}

func (s *Storage) PickRandom(ctx context.Context, userName string) (*storage.Page, error) {
	q := `SELECT url FROM pages WHERE user_name = ? ORDER BY RANDOM() LIMIT 1`

	row, err := s.db.QueryContext(ctx, q, userName)
	if err != nil {
		return nil, fmt.Errorf("can't get page: %w", err)
	}

	//получим данные в нужном формате
	var url string

	if err := row.Scan(&url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNoSavedPages
		}
	}

	return &storage.Page{URL: url, UserName: userName}, nil
}

func (s *Storage) Remove(ctx context.Context, p *storage.Page) error {
	q := `DELETE FROM pages WHERE url = ? AND user_name = ?`

	_, err := s.db.ExecContext(ctx, q, p.URL, p.UserName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("can't remove page: %w", err)
	}

	return nil
}

//IsExists check if page exist in storage.
func (s *Storage) IsExists(ctx context.Context, p *storage.Page) (bool, error) {
	q := `SELECT COUNT(*) FROM pages WHERE  url = ? AND user_name = ? `

	var count int
	if err := s.db.QueryRowContext(ctx, q, p.URL, p.UserName).Scan(&count); err != nil {
		return false, fmt.Errorf("can't check if page exist: %w", err)
	}

	return count > 0, nil
}

//инициализация Storage - создаст таблицу для хранения данных
func (s *Storage) Init(ctx context.Context) error {
	q := `CREATE TABLE IF NOT EXISTS pages (url TEXT, user_name TEXT)`

	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return fmt.Errorf("can't create table: %w", err)
	}

	return nil
}