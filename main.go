package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// from https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func openBookmark(b *Bookmark) {
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

func main() {
	// config, err := getConfig()
	db, err := m.newDB("./bm.db")
}
