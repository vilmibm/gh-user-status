package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	mutation := `mutation($emoji: String!, $message: String!) {
		changeUserStatus(input: {emoji: $emoji, message: $message}) {
			status {
				message
				emoji
			}
		}
	}`

	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		return fmt.Errorf("could not find gh. Is it installed? error: %w", err)
	}

	cmd := exec.Command(ghBin, "api", "graphql",
		"-f", fmt.Sprintf("query=%s", mutation),
		"-f", "emoji=:palm_tree:",
		"-f", "message=foobar")

	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run gh: %w", err)
	}

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
	s, err := apiStatus(opts.Login)
	if err != nil {
		return err
	}

	availability := ""
	if s.IndicatesLimitedAvailability {
		availability = "(availability is limited)"
	}
	fmt.Printf("%s %s %s\n",
		s.Emoji, // TODO try and map to unicode
		s.Message,
		availability)

	return nil
}

func apiStatus(login string) (*status, error) {
	key := "user"
	query := fmt.Sprintf(
		`query { user(login:"%s") { status { indicatesLimitedAvailability message emoji }}}`,
		login)
	if login == "" {
		key = "viewer"
		query = `query {viewer { status { indicatesLimitedAvailability message emoji }}}`
	}

	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		return nil, fmt.Errorf("could not find gh. Is it installed? error: %w", err)
	}
	var out bytes.Buffer
	cmd := exec.Command(ghBin, "api", "graphql", "-f", fmt.Sprintf("query=%s", query))
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run gh: %w", err)
	}

	resp := map[string]map[string]map[string]status{}

	err = json.Unmarshal(out.Bytes(), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize JSON: %w", err)
	}

	s, ok := resp["data"][key]["status"]
	if !ok {
		return nil, errors.New("failed to deserialize JSON")
	}

	return &s, nil
}

func main() {
	rc := rootCmd()
	rc.AddCommand(setCmd())
	rc.AddCommand(getCmd())

	if err := rc.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
