/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bookmarker/pkg"
	"fmt"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all bookmarks to file",
	Long:  ``,
	RunE:  Export,
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

	jsonName, _ := cmd.Flags().GetString("json")
	outputPath := fmt.Sprintf("./%s", jsonName)

	if jsonName != "" {
		if err = pkg.ExportToJSON(bList, outputPath); err != nil {
			return err
		}
	}

	// export to text by default
	if err = pkg.ExportToText(bList, "./output"); err != nil {
		return err
	}
	return nil
}
