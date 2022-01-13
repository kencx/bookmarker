package cmd

import (
	"bookmarker/pkg"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:     "export filename",
	Short:   "Export all bookmarks to file",
	Long:    `Export all bookmarks to file with given extension.`,
	Example: "bookmarker export bookmarks[.json|.html|.md]",
	Args:    cobra.MinimumNArgs(1),
	RunE:    Export,
}

func Export(cmd *cobra.Command, args []string) error {

	if err := validateExtension(args[0]); err != nil {
		return err
	}

	db, err := pkg.NewDB("bm.db") // TODO: replace with default path
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	s := pkg.Storage{Db: db}
	bList, err := s.GetAllBookmarks()
	if err != nil {
		return err
	}
	if len(bList) == 0 {
		return errors.New("no bookmarks found")
	}

	filename := args[0]
	extension := filepath.Ext(filename)
	if err = pkg.ExportBookmarks(bList, filename, extension); err != nil {
		return err
	}
	return nil
}

func validateExtension(filename string) error {
	extension := filepath.Ext(filename)
	extMap := map[string]bool{
		".json": true,
		".md":   true,
		".html": true,
		"":      true,
	}

	if !extMap[extension] {
		return fmt.Errorf("invalid extension: %s", extension)
	}
	return nil
}
