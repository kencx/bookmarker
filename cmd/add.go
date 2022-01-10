package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// implementation
	},
}

func Execute() {
	if err := addCmd.Execute(); err != nil {
		fmt.Errorf("%v", err)
	}
}
