package cmd

import (
	"bookmarker/pkg"
	"fmt"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:     "export [-j, --json] filename",
	Short:   "Export all bookmarks to file",
	Long:    ``,
	Example: "bookmarker export -j bookmarks.json",
	Args:    cobra.MinimumNArgs(1),
	RunE:    Export,
}

func Export(cmd *cobra.Command, args []string) error {

	db, err := pkg.NewDB("bm.db") // TODO: replace with default path
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	s := pkg.Storage{Db: db}
	bList, err := s.GetAllBookmarks()
	if err != nil {
		return err
	}

	jsonFileName, _ := cmd.Flags().GetString("json")

	if jsonFileName != "" {
		outputPath := fmt.Sprintf("./%s", jsonFileName)
		if err = pkg.ExportToJSON(bList, outputPath); err != nil {
			return err
		}
	} else {
		outputPath := fmt.Sprintf("./%s", args[0])
		if err = pkg.ExportToText(bList, outputPath); err != nil {
			return err
		}
	}
	return nil
}
