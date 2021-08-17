package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/safeexec"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	return &cobra.Command{
		Use: "user-status",
	}
}

type setOptions struct {
	Message string
	Limited bool
	Expiry  time.Duration
	Emoji   string
	OrgName string
}

func prompt(em emojiManager, opts *setOptions) error {
	emojiChoices := []string{}
	for _, e := range em.Emojis() {
		emojiChoices = append(emojiChoices, fmt.Sprintf("%s %s %s", string(e.codepoint), e.names, e.desc))
	}
	qs := []*survey.Question{
		{
			Name:     "status",
			Prompt:   &survey.Input{Message: "Status"},
			Validate: survey.Required,
		},
		{
			Name: "emoji",
			Prompt: &survey.Select{
				Message: "Emoji",
				Options: emojiChoices,
				Default: 147,
			},
		},
		{
			Name: "limited",
			Prompt: &survey.Confirm{
				Message: "Indicate limited availability?",
			},
		},
	}
	answers := struct {
		Status  string
		Emoji   int
		Limited bool
	}{}
	err := survey.Ask(qs, &answers)
	if err != nil {
		return err
	}

	opts.Message = answers.Status
	opts.Emoji = em.Emojis()[answers.Emoji].names[0]
	opts.Limited = answers.Limited

	return nil
}

func setCmd() *cobra.Command {
	opts := setOptions{}
	cmd := &cobra.Command{
		Use:   "set <status>",
		Short: "set your GitHub status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Message = args[0]
			}
			return runSet(opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Emoji, "emoji", "e", "thought_balloon", "Emoji for status")
	cmd.Flags().BoolVarP(&opts.Limited, "limited", "l", false, "Indicate limited availability")
	cmd.Flags().DurationVarP(&opts.Expiry, "expiry", "E", time.Duration(0), "Expire status after this duration")
	cmd.Flags().StringVarP(&opts.OrgName, "org", "o", "", "Limit status visibility to an organization")

	return cmd
}

func runSet(opts setOptions) error {
	em := newEmojiManager()
	if opts.Message == "" {
		err := prompt(em, &opts)
		if err != nil {
			return err
		}
	}
	// TODO org flag -- punted on this bc i have to resolve an org ID and it didn't feel worth it.
	mutation := `mutation($emoji: String!, $message: String!, $limited: Boolean!, $expiry: DateTime) {
		changeUserStatus(input: {emoji: $emoji, message: $message, limitedAvailability: $limited, expiresAt: $expiry}) {
			status {
				message
				emoji
			}
		}
	}`

	limited := "false"
	if opts.Limited {
		limited = "true"
	}

	expiry := "null"
	if opts.Expiry > time.Duration(0) {
		expiry = time.Now().Add(opts.Expiry).Format("2006-01-02T15:04:05-0700")
	}

	emoji := fmt.Sprintf(":%s:", opts.Emoji)

	cmdArgs := []string{
		"api", "graphql",
		"-f", fmt.Sprintf("query=%s", mutation),
		"-f", fmt.Sprintf("message=%s", opts.Message),
		"-f", fmt.Sprintf("emoji=%s", emoji),
		"-F", fmt.Sprintf("limited=%s", limited),
		"-F", fmt.Sprintf("expiry=%s", expiry),
	}

	out, _, err := gh(cmdArgs...)
	if err != nil {
		return err
	}
	type response struct {
		Data struct {
			ChangeUserStatus struct {
				Status status
			}
		}
	}
	var resp response
	err = json.Unmarshal(out.Bytes(), &resp)
	if err != nil {
		return fmt.Errorf("failed to deserialize JSON: %w", err)
	}

	if resp.Data.ChangeUserStatus.Status.Emoji != emoji {
		return errors.New("failed to set status. Perhaps try another emoji")
	}

	msg := fmt.Sprintf("✓ Status set to %s %s", emoji, opts.Message)
	fmt.Println(em.ReplaceAll(msg))

	return nil
}

type getOptions struct {
	Login string
}

func getCmd() *cobra.Command {
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
	em := newEmojiManager()

	s, err := apiStatus(opts.Login)
	if err != nil {
		return err
	}

	availability := ""
	if s.IndicatesLimitedAvailability {
		availability = "(availability is limited)"
	}
	msg := fmt.Sprintf("%s %s %s", s.Emoji, s.Message, availability)

	fmt.Println(em.ReplaceAll(msg))

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

	args := []string{"api", "graphql", "-f", fmt.Sprintf("query=%s", query)}
	sout, _, err := gh(args...)
	if err != nil {
		return nil, err
	}

	resp := map[string]map[string]map[string]status{}

	err = json.Unmarshal(sout.Bytes(), &resp)
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

// gh shells out to gh, returning STDOUT/STDERR and any error
func gh(args ...string) (sout, eout bytes.Buffer, err error) {
	ghBin, err := safeexec.LookPath("gh")
	if err != nil {
		err = fmt.Errorf("could not find gh. Is it installed? error: %w", err)
		return
	}

	cmd := exec.Command(ghBin, args...)
	cmd.Stderr = &eout
	cmd.Stdout = &sout

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to run gh. error: %w, stderr: %s", err, eout.String())
		return
	}

	return
}
