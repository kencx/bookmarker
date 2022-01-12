package pkg

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	driver   = "sqlite3"
	dbparent = "bookmarker"
	filename = "bm.db"
)

type Bookmark struct {
	Id   int    `json:"-"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Storage struct {
	Db *sql.DB
}

func NewDB(path string) (*sql.DB, error) {
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

	stmt, err := s.Db.Prepare("INSERT INTO bookmarks (name, url) VALUES (?, ?)")
	if err != nil {
		return -1, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(b.Name, b.Url)
	if err != nil {
		return -1, fmt.Errorf("failed to add row: %w", err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("failed to get lastId: %w", err)
	}
	b.Id = int(lastId)
	log.Printf("bookmark %d added successfully", b.Id)
	return int(lastId), nil
}

func (s *Storage) GetBookmarkById(id int) (Bookmark, error) {

	stmt, err := s.Db.Prepare("SELECT id, name, url FROM bookmarks WHERE id = ?")
	if err != nil {
		return Bookmark{}, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	var b Bookmark
	if err = stmt.QueryRow(id).Scan(&b.Id, &b.Name, &b.Url); err != nil {
		if err != sql.ErrNoRows {
			return Bookmark{}, fmt.Errorf("failed to query row: %w", err)
		}
	}
	return b, nil
}

func (s *Storage) GetAllBookmarks() ([]Bookmark, error) {

	stmt, err := s.Db.Prepare("SELECT name, url FROM bookmarks")
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
		if err = rows.Scan(&b.Name, &b.Url); err != nil {
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

	stmt, err := s.Db.Prepare("DELETE FROM bookmarks WHERE id = ?")
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
	res, err := s.Db.Exec("DELETE FROM bookmarks")
	if err != nil {
		return -1, fmt.Errorf("failed to delete all bookmarks: %w", err)
	}
	log.Printf("dropped bookmarks table successfully")
	return res.RowsAffected()
}

func (s *Storage) UpdateBookmark(id int, b *Bookmark) (int64, error) {

	stmt, err := s.Db.Prepare("UPDATE bookmarks SET name = ?, url = ? WHERE id = ?")
	if err != nil {
		return -1, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(b.Name, b.Url, id)
	if err != nil {
		return -1, fmt.Errorf("failed to update row: %w", err)
	}
	log.Printf("bookmark %d updated successfully", id)
	return res.RowsAffected()
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

func CreateDBDir() error {

	basedir, err := GetBaseDir()
	if err != nil {
		return err
	}
	dbpath := filepath.Join(basedir, dbparent)
	fullpath := filepath.Join(dbpath, filename)

	_, err = os.Stat(fullpath)
	if errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(dbpath, 0755)
		if err != nil {
			return fmt.Errorf("error creating parent dir: %v", err)
		}
	} else {
		return fmt.Errorf("error determining file structure: %v", err)
	}
	return nil
}

// from https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func (b *Bookmark) OpenBookmark() error {
	url := b.Url
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

func GetNameFromUrl(url string) (string, error) {
	return "", nil
}

func ToText(bList []Bookmark) (string, error) {
	var sb strings.Builder

	for _, b := range bList {
		temp := fmt.Sprintf("%d. %s - %s\n", b.Id, b.Name, b.Url)
		_, err := sb.WriteString(temp)
		if err != nil {
			return "", fmt.Errorf("could not parse output: %w", err)
		}
	}

	return sb.String(), nil
}

func ToJSON(bList []Bookmark) ([]byte, error) {
	result, err := json.MarshalIndent(bList, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %w", err)
	}
	return result, nil
}

func ExportToText(bList []Bookmark, path string) error {
	result, err := ToText(bList)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, []byte(result), 0700)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil

}

func ExportToJSON(bList []Bookmark, path string) error {
	result, err := ToJSON(bList)
	if err != nil {
		return err
	}

	path = fmt.Sprintf("%s.json", path)
	err = os.WriteFile(path, result, 0700)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}
