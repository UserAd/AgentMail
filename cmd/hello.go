package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helloName string

var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "Say hello to someone",
	Long:  `Say hello to a person. You can specify a custom name using the --name flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Hello, %s!\n", helloName)
	},
}

func init() {
	rootCmd.AddCommand(helloCmd)
	helloCmd.Flags().StringVarP(&helloName, "name", "n", "World", "Name to greet")
}
