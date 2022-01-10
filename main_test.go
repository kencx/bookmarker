package main

import (
	"fmt"
	"os"
	"testing"
)

const (
	path = "./bm.db"
)

var s *Storage

func TestMain(m *testing.M) {

	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {

	db, err := newDB(path)
	if err != nil {
		return -1, fmt.Errorf("could not connect to database: %w", err)
	}

	defer func() {
		db.Exec("DROP TABLE bookmarks")
		db.Close()
	}()

	s = &Storage{db: db}

	return m.Run(), nil
}

func TestConnection(t *testing.T) {
	if err := s.db.Ping(); err != nil {
		t.Errorf("database connection failed: %w", err)
	}
}

func TestAddBookmark(t *testing.T) {

	tb := []struct {
		bookmark Bookmark
		want     int64
	}{
		{Bookmark{name: "google", url: "google.com"}, 1},
		{Bookmark{name: "reddit", url: "reddit.com"}, 2},
	}

	t.Run("Add bookmark", func(t *testing.T) {

		for _, tt := range tb {
			got, err := s.AddBookmark(&tt.bookmark)
			if err != nil {
				t.Errorf("AddBookmark failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}

			t.Cleanup(func() {
				if err = s.DeleteAllBookmarks(); err != nil {
					t.Fatalf("cleanup failed: %v", err)
				}
			})
		}
	})
}

func TestGetAllBookmarks(t *testing.T) {

	t.Run("Get all bookmarks", func(t *testing.T) {

		tb := []Bookmark{
			Bookmark{name: "google", url: "google.com"},
			Bookmark{name: "reddit", url: "reddit.com"},
		}
		for _, tt := range tb {
			_, err := s.AddBookmark(&tt)
			if err != nil {
				t.Errorf("setup failed: %v", err)
			}
		}

		bList, err := s.GetAllBookmarks()
		if err != nil {
			t.Fatalf("GetAllBookmarks failed: %v", err)
		}
		if len(bList) != len(tb) {
			t.Errorf("got %q, want %q", len(bList), len(tb))
		}

		t.Cleanup(func() {
			if err = s.DeleteAllBookmarks(); err != nil {
				t.Fatalf("cleanup failed: %v", err)
			}
		})
	})
}

func TestGetBookmarkById(t *testing.T) {

}

func TestDeleteBookmark(t *testing.T) {

}

func TestDeleteAllBookmarks(t *testing.T) {

}

func TestUpdateBookmark(t *testing.T) {

}
