package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

const (
	driver   = "sqlite3"
	dbparent = "bookmarker"
	filename = "bm.db"
)

type Bookmark struct {
	id   int
	name string
	url  string
}

type Storage struct {
	db *sql.DB
}

func newDB(path string) (*sql.DB, error) {
	db, err := sql.Open(driver, path)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
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

func (s *Storage) AddBookmark(b *Bookmark) (int, error) {

	stmt, err := s.db.Prepare("INSERT INTO bookmarks (name, url) VALUES (?, ?)")
	if err != nil {
		return -1, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(b.name, b.url)
	if err != nil {
		return -1, fmt.Errorf("failed to add row: %w", err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("failed to get lastId: %w", err)
	}
	b.id = int(lastId)
	log.Printf("bookmark %d added successfully", b.id)
	return int(lastId), nil
}

func (s *Storage) GetBookmarkById(id int) (Bookmark, error) {

	stmt, err := s.db.Prepare("SELECT id, name, url FROM bookmarks WHERE id = ?")
	if err != nil {
		return Bookmark{}, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	var b Bookmark
	if err = stmt.QueryRow(id).Scan(&b.id, &b.name, &b.url); err != nil {
		if err != sql.ErrNoRows {
			return Bookmark{}, fmt.Errorf("failed to query row: %w", err)
		}
	}
	return b, nil
}

func (s *Storage) GetAllBookmarks() ([]Bookmark, error) {

	stmt, err := s.db.Prepare("SELECT name, url FROM bookmarks")
	if err != nil {
		return nil, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %w", err)
	}
	defer rows.Close()

	var bList []Bookmark
	for rows.Next() {
		var b Bookmark
		if err = rows.Scan(&b.name, &b.url); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		bList = append(bList, b)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan rows: %w", err)
	}
	return bList, nil
}

func (s *Storage) DeleteBookmark(id int) (int64, error) {

	stmt, err := s.db.Prepare("DELETE FROM bookmarks WHERE id = ?")
	if err != nil {
		return -1, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return -1, fmt.Errorf("failed to delete row: %w", err)
	}
	log.Printf("bookmark deleted successfully")
	return res.RowsAffected()
}

func (s *Storage) DeleteAllBookmarks() (int64, error) {
	res, err := s.db.Exec("DELETE FROM bookmarks")
	if err != nil {
		return -1, fmt.Errorf("failed to delete all bookmarks: %w", err)
	}
	log.Printf("dropped bookmarks table successfully")
	return res.RowsAffected()
}

func (s *Storage) UpdateBookmark(id int, b *Bookmark) (int64, error) {

	stmt, err := s.db.Prepare("UPDATE bookmarks SET name = ?, url = ? WHERE id = ?")
	if err != nil {
		return -1, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(b.name, b.url, id)
	if err != nil {
		return -1, fmt.Errorf("failed to update row: %w", err)
	}
	log.Printf("bookmark %d updated successfully", id)
	return res.RowsAffected()
}

// from https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func (b *Bookmark) OpenBookmark() error {
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
		err = fmt.Errorf("unsupported platform") // replace with defined error
	}
	if err != nil {
		return fmt.Errorf("failed to open url: %w", err)
	}
	return nil
}

func GetBaseDir() (string, error) {
	var err error
	var path string

	switch runtime.GOOS {
	case "windows":
		path = os.Getenv("APPDATA")
	case "linux":
		// check config for custom dir
		path = os.Getenv("XDG_DATA_HOME")
		if path == "" {
			homepath := os.Getenv("HOME")
			path = filepath.Join(homepath, ".local", "share")
		}
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return path, err
}

func (s *Storage) importBookmarksJSON(path string) {

}

func (s *Storage) importBookmarksHTML(path string) {

}

// export to json, html, md
func (s *Storage) exportBookmarksJSON(path string) {

}

func main() {

	// get config file

	basedir, err := GetBaseDir()
	if err != nil {
		log.Fatal(err)
	}
	dbpath := filepath.Join(basedir, dbparent)
	fullpath := filepath.Join(dbpath, filename)

	_, err = os.Stat(fullpath)
	if errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(dbpath, 0755)
		if err != nil {
			log.Fatalf("Error creating parent dir: %v", err)
		}
	} else {
		log.Fatalf("Error determining file structure: %v", err)
	}

	db, err := newDB(fullpath)
	if err != nil {
		log.Fatal(err)
	}

	s := &Storage{db: db}

	if err = s.db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("database connection successful")

	// commands

	// checking seeded data is cleared
	bList, err := s.GetAllBookmarks()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Number of active rows: ", len(bList))
}
