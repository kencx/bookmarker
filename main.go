package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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
		return nil, fmt.Errorf("failed to establish connection: %v", err)
	}
	if err = createTable(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTable(db *sql.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		url TEXT NOT NULL
	);`

	_, err := db.Exec(stmt)
	if err != nil {
		return fmt.Errorf("failed to create table bookmarks: %w", err)
	}
	return nil
}

func (s *Storage) AddBookmark(b *Bookmark) (int64, error) {

	stmt, err := s.db.Prepare("INSERT INTO bookmarks (name, url) VALUES (?, ?)")
	if err != nil {
		return -1, fmt.Errorf("prepare stmt failed: %v", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(b.name, b.url)
	if err != nil {
		return -1, fmt.Errorf("failed to add row: %v", err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("failed to get lastId: %v", err)
	}
	b.id = lastId
	log.Printf("bookmark %d added successfully", b.id)
	return lastId, nil
}

func (s *Storage) GetBookmarkById(id int64) (*Bookmark, error) {

	stmt, err := s.db.Prepare("SELECT (name, url) FROM bookmarks WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("prepare stmt failed: %v", err)
	}
	defer stmt.Close()

	var b *Bookmark
	if err = stmt.QueryRow(id).Scan(&b.name, &b.url); err != nil {
		if err == sql.ErrNoRows {
			// no error
		} else {
			return nil, fmt.Errorf("failed to query row: %v", err)
		}
	}
	return b, nil
}

func (s *Storage) GetAllBookmarks() ([]Bookmark, error) {

	stmt, err := s.db.Prepare("SELECT name, url FROM bookmarks")
	if err != nil {
		return nil, fmt.Errorf("prepare stmt failed: %v", err)
	}
	defer stmt.Close()

	var bList []Bookmark
	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Bookmark
		if err = rows.Scan(&b.name, &b.url); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		bList = append(bList, b)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan rows: %v", err)
	}
	return bList, nil
}

func (s *Storage) DeleteBookmark(id int64) error {

	stmt, err := s.db.Prepare("DELETE FROM bookmarks WHERE id = ?")
	if err != nil {
		return fmt.Errorf("prepare stmt failed: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to delete row: %v", err)
	}
	log.Printf("bookmark %d deleted successfully", id)
	return nil
}

func (s *Storage) DeleteAllBookmarks() error {
	_, err := s.db.Exec("DELETE FROM bookmarks")
	if err != nil {
		return fmt.Errorf("failed to drop bookmarks: %v", err)
	}
	log.Printf("dropped bookmarks table successfully")
	return nil
}

func (s *Storage) UpdateBookmark(id int64, b *Bookmark) error {

	stmt, err := s.db.Prepare("UPDATE bookmarks SET name = ?, url = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("prepare stmt failed: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(b.name, b.url, id)
	if err != nil {
		return fmt.Errorf("failed to update row: %v", err)
	}
	log.Printf("bookmark %d updated successfully", id)
	return nil
}

func (b *Bookmark) openBookmark() {
	url := b.url
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatalf("failed to open url: %v", err)
	}
}

// from https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func GetDBDir() (string, error) {
	var err error
	var fullpath string

	switch runtime.GOOS {
	case "windows":
		fullpath = os.Getenv("APPDATA")
	case "linux":
		fullpath = os.Getenv("XDG_DATA_HOME")
		if fullpath == "" {
			homepath := os.Getenv("HOME")
			fullpath = filepath.Join(homepath, ".local", "share")
		}
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return fullpath, err
}

func main() {
	db, err := newDB("./bm.db")
	if err != nil {
		log.Fatal(err)
	}

	s := &Storage{db: db}

	if err = s.db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("database connection successful")

	bList, err := s.GetAllBookmarks()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(bList))
}
