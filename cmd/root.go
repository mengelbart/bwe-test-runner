package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bwetest",
	Short: "Bandwidth estimation test runner",
}

func Execute() {
	rootCmd.Execute()
}
