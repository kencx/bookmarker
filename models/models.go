package models

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Bookmark struct {
	id   int64
	name string
	url  string
}

type Storage struct {
	db *sql.DB
}

func newDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}

	return db, nil
}

func createTable(db *sql.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY NOT NULL AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
	);`

	_, err := db.Exec(stmt)
	if err != nil {
		return fmt.Errorf("failed to create table bookmarks")
	}
	return nil
}

func (s *Storage) AddBookmark(b *Bookmark) error {

	stmt, err := s.db.Prepare("INSERT INTO bookmarks (name, url) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(b.name, b.url)
	if err != nil {
		return fmt.Errorf("failed to add row: %w", err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get lastId: %w", err)
	}
	b.id = lastId
	log.Printf("bookmark %d added successfully", b.id)
	return nil
}

func (s *Storage) GetBookmark(id int64) (*Bookmark, error) {

	stmt, err := s.db.Prepare("SELECT (name, url) FROM bookmarks WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	var b *Bookmark
	err = stmt.QueryRow(id).Scan(&b.name, &b.url)
	if err != nil {
		if err == sql.ErrNoRows {
			// no error
		} else {
			return nil, fmt.Errorf("failed to query row: %w", err)
		}
	}
	return b, nil
}

func (s *Storage) GetAllBookmarks() ([]Bookmark, error) {

	stmt, err := s.db.Prepare("SELECT (name, url) FROM bookmarks")
	if err != nil {
		return nil, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	var bList []Bookmark
	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Bookmark
		err = rows.Scan(&b.name, &b.url)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		bList = append(bList, b)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to scan rows: %w", err)
	}
	return bList, nil
}

func (s *Storage) DeleteBookmark(id int64) error {

	stmt, err := s.db.Prepare("DELETE FROM bookmarks WHERE id = ?")
	if err != nil {
		return fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to delete row: %w", err)
	}
	log.Printf("bookmark %d deleted successfully", id)
	return nil
}

func (s *Storage) UpdateBookmark(id int64, b *Bookmark) error {

	stmt, err := s.db.Prepare("UPDATE bookmarks SET name = ?, url = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(b.name, b.url, id)
	if err != nil {
		return fmt.Errorf("failed to update row: %w", err)
	}
	log.Printf("bookmark %d updated successfully", id)
	return nil
}
