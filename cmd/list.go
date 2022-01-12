package cmd

import (
	"bookmarker/pkg"
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var result string
var listCmd = &cobra.Command{
	Use:     "list [-j, --json]",
	Short:   "List bookmarks",
	Long:    ``,
	Example: "bookmarker list -j",
	RunE:    list,
}

func list(cmd *cobra.Command, args []string) error {

	db, err := pkg.NewDB("bm.db") // TODO: replace with default path
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	s := pkg.Storage{Db: db}
	bList, err := s.GetAllBookmarks()
	if err != nil {
		return err
	}

	printJson, _ := cmd.Flags().GetBool("json")
	if printJson {
		json, err := pkg.ToJSON(bList)
		result = string(json)
		if err != nil {
			return err
		}
	} else {
		result, err = pkg.ToText(bList)
		if err != nil {
			return err
		}
	}

	fmt.Println(result)
	return nil
}
