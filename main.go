package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cli/safeexec"
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
		Use:   "set <status>",
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
	/*
			# We'll get you started with a simple query showing your username!
		mutation ($status: ChangeUserStatusInput!) {
		    changeUserStatus(input: $status) {
		      status {
		        emoji
		        expiresAt
		        limitedAvailability: indicatesLimitedAvailability
		        message
		      }
		    }
		  }*/
	fmt.Printf("set %s\n", opts.Status)
	return nil
}

type getOptions struct {
	Login string
}

func getCmd() *cobra.Command {
	// TODO get arbitrary user
	// TODO get current user
	return &cobra.Command{
		Use:   "get [<username>]",
		Short: "get a GitHub user's status or your own",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := getOptions{}
			if len(args) > 0 {
				opts.Login = args[0]
			}
			return runGet(opts)
		},
	}
}

type status struct {
	IndicatesLimitedAvailability bool
	Message                      string
	Emoji                        string
}

func runGet(opts getOptions) error {
	if opts.Login == "" {
		panic("TODO")
	}
	s, err := apiStatus(opts.Login)
	if err != nil {
		return err
	}

	// TODO
	fmt.Println(s)

	return nil
}

func apiStatus(login string) (status, error) {
	query := fmt.Sprintf(`query getUserStatus {
		user(login:"%s") {
			status {
				indicatesLimitedAvailability
				message
				emoji
		}}}`, login)

	// TODO
	fmt.Println(query)
	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		return status{}, err
	}
	// TODO gh api just opaquely returning exit status 1, why? is there an
	// escaping problem? gh api is running fine manually.
	// the env thing is stupid and didn't help
	queryValue := fmt.Sprintf("STATUS_QUERY=query='%s'", query)
	cmd := exec.Command(ghBin, "api", "graphql", "-f", "$STATUS_QUERY")
	cmd.Env = append(os.Environ(), queryValue)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("DBG %#v\n", cmd)
		fmt.Printf("DBG %#v\n", err)
		return status{}, err
	}

	return status{}, nil
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
