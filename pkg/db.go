package pkg

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mattn/go-sqlite3"
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
		url TEXT NOT NULL UNIQUE
	);`

	_, err := db.Exec(stmt)
	if err != nil {
		return fmt.Errorf("create table bookmarks failed: %w", err)
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
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.Code == sqlite3.ErrConstraint && sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return -1, fmt.Errorf("url already exists: %s", b.Url)
			}
		} else {
			return -1, fmt.Errorf("add row failed: %w", err)
		}
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

	stmt, err := s.Db.Prepare("SELECT * FROM bookmarks WHERE id = ?")
	if err != nil {
		return Bookmark{}, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	var b Bookmark
	err = stmt.QueryRow(id).Scan(&b.Id, &b.Name, &b.Url)
	if err == sql.ErrNoRows {
		return Bookmark{}, nil
	} else if err != nil {
		return Bookmark{}, fmt.Errorf("query row failed: %w", err)
	}

	return b, nil
}

func (s *Storage) GetAllBookmarks() ([]Bookmark, error) {

	stmt, err := s.Db.Prepare("SELECT * FROM bookmarks")
	if err != nil {
		return nil, fmt.Errorf("prepare stmt failed: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("query row failed: %w", err)
	}
	defer rows.Close()

	var bList []Bookmark
	for rows.Next() {
		var b Bookmark
		if err = rows.Scan(&b.Id, &b.Name, &b.Url); err != nil {
			return nil, fmt.Errorf("scan row failed: %w", err)
		}
		bList = append(bList, b)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan rows failed: %w", err)
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
		return -1, fmt.Errorf("delete row failed: %w", err)
	}
	log.Printf("bookmark deleted successfully")
	return res.RowsAffected()
}

func (s *Storage) DeleteAllBookmarks() (int64, error) {
	res, err := s.Db.Exec("DELETE FROM bookmarks")
	if err != nil {
		return -1, fmt.Errorf("delete all bookmarks failed: %w", err)
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
		return -1, fmt.Errorf("update row failed: %w", err)
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

	getTime := time.Now()
	var title string

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Get(url)

	if err != nil {
		return "", fmt.Errorf("failed to get url: %w", err)
	}
	defer resp.Body.Close()
	elapsed := time.Since(getTime)
	log.Printf("HTTP get took %s", elapsed)

	if resp.StatusCode == http.StatusOK {
		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			return "", fmt.Errorf("failed to parse response: %w", err)
		}
		title = doc.Find("title").Text()
		return title, nil
	}
	return "", errors.New("title not found")
}

func ToText(bList []Bookmark) (string, error) {
	var sb strings.Builder

	for _, b := range bList {
		row := fmt.Sprintf("%d. %s - %s\n", b.Id, b.Name, b.Url)

		_, err := sb.WriteString(row)
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

func ToMarkdown(bList []Bookmark) (string, error) {
	var sb strings.Builder

	for _, b := range bList {
		row := fmt.Sprintf("- [%s](%s)\n", b.Name, b.Url) // TODO default name if empty
		_, err := sb.WriteString(row)
		if err != nil {
			return "", fmt.Errorf("could not parse output: %w", err)
		}
	}
	return sb.String(), nil
}

func ExportBookmarks(bList []Bookmark, path string, format string) error {
	var result []byte
	var err error

	switch format {
	case ".json":
		result, err = ToJSON(bList)
		if err != nil {
			return err
		}
	case ".md":
		temp, err := ToMarkdown(bList)
		if err != nil {
			return err
		}
		result = []byte(temp)
	default:
		temp, err := ToText(bList)
		if err != nil {
			return err
		}
		result = []byte(temp)
	}

	err = os.WriteFile(path, result, 0700)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
