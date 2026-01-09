package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of AgentMail",
	Long:  `Display the current version of AgentMail CLI application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("AgentMail version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// GetVersion returns the current version
func GetVersion() string {
	return version
}
