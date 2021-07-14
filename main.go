package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	return &cobra.Command{
		Use: "user-status",
	}
}

func setCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "set your GitHub status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("TODO sc")
			return nil
		},
	}
}

func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "get a GitHub user's status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("TODO gc")
			return nil
		},
	}
}

func main() {
	rc := rootCmd()
	sc := setCmd()
	gc := getCmd()
	rc.AddCommand(sc)
	rc.AddCommand(gc)

	if err := rc.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
