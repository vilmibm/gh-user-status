package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	return &cobra.Command{
		Use: "user-status",
	}
}

type setOptions struct {
	Status  string
	Limited bool
	Expiry  time.Duration
	Emoji   string
	OrgName string
}

func setCmd() *cobra.Command {
	opts := setOptions{}
	cmd := &cobra.Command{
		Use:   "set [<status>]",
		Short: "set your GitHub status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Status = args[0]
			return runSet(opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Emoji, "emoji", "e", "", "Emoji for status")
	cmd.Flags().BoolVarP(&opts.Limited, "limited", "l", false, "Indicate limited availability")
	cmd.Flags().DurationVarP(&opts.Expiry, "expiry", "E", time.Duration(0), "Expire status after this duration")
	cmd.Flags().StringVarP(&opts.OrgName, "org", "o", "", "Limit status visibility to an organization")

	return cmd
}

func runSet(opts setOptions) error {
	// TODO limited flag
	// TODO expiry flag
	// TODO emoji flag
	// TODO org flag
	fmt.Printf("set %s\n", opts.Status)
	return nil
}

func getCmd() *cobra.Command {
	// TODO get arbitrary user
	// TODO get current user
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
