package pkg

import (
	"fmt"
	"os"
	"testing"
)

const (
	path = "./bm.db"
)

var (
	s             *Storage
	testBookmark  = Bookmark{Name: "google", Url: "google.com"}
	testBookmark2 = Bookmark{Name: "reddit", Url: "reddit.com"}
	testList      = []Bookmark{testBookmark, testBookmark2}
)

func TestMain(m *testing.M) {

	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {

	db, err := NewDB(path)
	if err != nil {
		return -1, fmt.Errorf("could not connect to database: %w", err)
	}
	defer func() {
		db.Exec("DROP TABLE bookmarks")
		db.Close()
	}()
	s = &Storage{Db: db}
	return m.Run(), nil
}

func TestConnection(t *testing.T) {
	if err := s.Db.Ping(); err != nil {
		t.Errorf("database connection failed: %v", err)
	}
}

func TestAddBookmark(t *testing.T) {

	t.Run("Add bookmark", func(t *testing.T) {

		got, err := s.AddBookmark(&testBookmark)
		if err != nil {
			t.Errorf("unexpected error in AddBookmark: %v", err)
		}
		assertEquals(t, got, 1)
		teardown(t)
	})
}

func TestGetAllBookmarks(t *testing.T) {

	t.Run("Get all bookmarks", func(t *testing.T) {

		seedMultiple(t)
		bList, err := s.GetAllBookmarks()
		if err != nil {
			t.Fatalf("GetAllBookmarks failed: %v", err)
		}

		assertEquals(t, len(bList), len(testList))
		for i := range bList {
			assertBookmarkEquals(t, bList[i], testList[i])
		}
		teardown(t)
	})
}

func TestGetBookmarkById(t *testing.T) {

	t.Run("get existing bookmark by id", func(t *testing.T) {
		id := seedOne(t)
		got, err := s.GetBookmarkById(id)
		if err != nil {
			t.Fatalf("GetBookmarkById failed: %v", err)
		}
		want := testBookmark
		assertBookmarkEquals(t, got, want)
		teardown(t)
	})

	t.Run("get non-existent bookmark", func(t *testing.T) {
		got, err := s.GetBookmarkById(1000)
		if err != nil {
			t.Fatalf("GetBookmarkById failed: %v", err)
		}
		want := Bookmark{}
		if got != want {
			t.Fatalf("non-existent bookmark found: %q", got)
		}
	})
}

func TestDeleteBookmark(t *testing.T) {
	t.Run("delete bookmark", func(t *testing.T) {
		id := seedOne(t)
		rows, err := s.DeleteBookmark(id)
		if err != nil {
			t.Fatalf("DeleteBookmark failed: %v", err)
		}
		if rows != 1 {
			t.Fatalf("wrong number of bookmarks deleted: %d/1", rows)
		}
		teardown(t)
	})

	t.Run("delete 1 bookmark only out of 2", func(t *testing.T) {
		idList := seedMultiple(t)
		rows, err := s.DeleteBookmark(idList[0])
		if err != nil {
			t.Fatalf("DeleteBookmark failed: %v", err)
		}
		if rows != 1 {
			t.Fatalf("wrong number of bookmarks deleted: %d/1", rows)
		}
		bList, err := s.GetAllBookmarks()
		if err != nil {
			t.Errorf("get bookmarks failed: %v", err)
		}
		if len(bList) != len(idList)-1 {
			t.Fatalf("wrong number of bookmarks remaining: %d/%d", len(bList), len(idList)-1)
		}
		teardown(t)
	})

	t.Run("delete non-existent bookmark", func(t *testing.T) {
		rows, err := s.DeleteBookmark(1000)
		if err != nil {
			t.Fatalf("DeleteBookmark failed: %v", err)
		}
		if rows != 0 {
			t.Fatalf("wrong number of bookmarks deleted: %d/0", rows)
		}
		teardown(t)
	})
}

func TestDeleteAllBookmarks(t *testing.T) {

	idList := seedMultiple(t)
	rows, err := s.DeleteAllBookmarks()
	if err != nil {
		t.Fatalf("DeleteAllBookmarks failed: %v", err)
	}
	if int(rows) != len(idList) {
		t.Fatalf("wrong number of bookmarks deleted: %d/%d", int(rows), len(idList))
	}
	bList, err := s.GetAllBookmarks()
	if err != nil {
		t.Errorf("get bookmarks failed: %v", err)
	}
	if len(bList) != 0 {
		t.Fatalf("wrong number of bookmarks remaining: %d/0", len(bList))
	}
	teardown(t)
}

func TestUpdateBookmark(t *testing.T) {

	id := seedOne(t)
	newTestBookmark := &Bookmark{Name: "google search", Url: "google.com"}
	rows, err := s.UpdateBookmark(id, newTestBookmark)
	if err != nil {
		t.Fatalf("UpdateBookmark failed: %v", err)
	}
	if rows != 1 {
		t.Fatalf("wrong number of bookmarks updated: %d/1", rows)
	}
	b, err := s.GetBookmarkById(id)
	if err != nil {
		t.Errorf("get bookmarks failed: %v", err)
	}
	assertBookmarkEquals(t, b, *newTestBookmark)
	teardown(t)
}

func TestGetBaseDir(t *testing.T) {

	t.Run("linux", func(t *testing.T) {
		got, err := GetBaseDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "/home/kenc/.local/share"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

func seedOne(t testing.TB) int {
	t.Helper()
	id, err := s.AddBookmark(&testBookmark)
	if err != nil {
		t.Errorf("seed database failed: %v", err)
	}
	return id
}

func seedMultiple(t testing.TB) []int {
	t.Helper()
	var idList []int
	for _, tt := range testList {
		id, err := s.AddBookmark(&tt)
		if err != nil {
			t.Errorf("seed database failed: %v", err)
		}
		idList = append(idList, id)
	}
	return idList
}

func teardown(t testing.TB) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := s.DeleteAllBookmarks(); err != nil {
			t.Fatalf("cleanup failed: %v", err)
		}
	})
}

func assertEquals(t testing.TB, got int, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertBookmarkEquals(t testing.TB, got Bookmark, want Bookmark) {
	t.Helper()
	if got.Name != want.Name {
		t.Errorf("got %q, want %q", got.Name, want.Name)
	}
	if got.Url != want.Url {
		t.Errorf("got %q, want %q", got.Url, want.Url)
	}
}
