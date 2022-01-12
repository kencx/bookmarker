package cmd

import (
	"bookmarker/pkg"
	"errors"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

const url_regexp = `[(http(s)?):\/\/(www\.)?a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&\/\/=]*)`

var bm *pkg.Bookmark
var addCmd = &cobra.Command{
	Use:     "add url [-n] name",
	Short:   "Add new bookmark",
	Example: "bookmarker add google.com -n google",
	Args:    validateAdd,
	RunE:    add,
}

func validateAdd(cmd *cobra.Command, args []string) error {

	if len(args) < 1 {
		return errors.New("missing url")
	}

	// url validation
	re := regexp.MustCompile(url_regexp)
	match := re.MatchString(args[0])
	if !match {
		return fmt.Errorf("invalid url format: %s", args[0])
	}

	// check if url already exists in db
	return nil
}

func add(cmd *cobra.Command, args []string) error {

	bmName, _ := cmd.Flags().GetString("name")
	if bmName != "" {
		bm = &pkg.Bookmark{Url: args[0], Name: args[1]}
	} else {
		name, err := pkg.GetNameFromUrl(args[0])
		if err != nil {
			return fmt.Errorf("failed to get name from url: %w", err)
		}
		bm = &pkg.Bookmark{Url: args[0], Name: name}
	}

	db, err := pkg.NewDB("bm.db")
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	s := pkg.Storage{Db: db}
	_, err = s.AddBookmark(bm)
	if err != nil {
		return err
	}

	return nil
}
