package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bookmarker",
	Short: "A command-line bookmark manager",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(exportCmd)

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test.yaml)")

	// local flags
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	addCmd.Flags().StringP("name", "n", "", "Name of bookmark")

	listCmd.Flags().BoolP("json", "j", false, "List in JSON format")

	exportCmd.Flags().StringP("json", "j", "", "Export to JSON file")
}
