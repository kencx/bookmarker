package models

import (
	"fmt"
	"log"
	"os"
	"testing"
)

const (
	path    = "./bm.db"
	testBM  = Bookmark{0, "google", "google.com"}
	testBM2 = Bookmark{0, "reddit", "reddit.com"}
)

func TestMain(m *testing.M) {

	code, err := run(m)
	if err != nil {
		log.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {

	db, err := initDB(path)
	if err != nil {
		return -1, fmt.Errorf("could not connect to database: %w", err)
	}

	defer func() {
		db.Exec("DELETE FROM bookmarks")
		db.Close()
	}()

	return m.Run(), nil
}

func TestCreate(t *testing.T) {
	CreateTable(db)
}

func TestCreateBookmark(t *testing.T) {

	AddBookmark(db, testBM)
	AddBookmark(db, testBM2)

}

func TestGetBookmark(t *testing.T) {
	result := GetBookmark(db, 1)
	fmt.Printf("%#v", result)
}

func TestGetAllBookmarks(t *testing.T) {
	resultList := GetAllBookmarks(db)
	fmt.Printf("%#v", resultList)
}

func TestDeleteBookmark(t *testing.T) {
	DeleteBookmark(db, 1)
}
