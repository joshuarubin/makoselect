package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type app struct {
	id    int64
	flags *flag.FlagSet
}

func newApp() *app {
	var a app
	a.flags = flag.NewFlagSet("", flag.ContinueOnError)
	a.flags.Int64Var(&a.id, "id", -1, "notification id")
	return &a
}

func (a *app) getNotifications(ctx context.Context) (*Notifications, error) {
	cmd := exec.Command("makoctl", "list")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	var notifications Notifications
	done := make(chan struct{})
	var decodeErr error
	go func() {
		defer close(done)
		decodeErr = json.NewDecoder(stdout).Decode(&notifications)
	}()

	if err = cmd.Run(); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		if decodeErr != nil {
			return nil, decodeErr
		}
	}

	return &notifications, nil
}

func (a *app) dismiss() error {
	return exec.Command(
		"makoctl",
		"dismiss",
		"-n", fmt.Sprintf("%d", a.id),
	).Run()
}

func (a *app) invoke(name string) error {
	return exec.Command(
		"makoctl",
		"invoke",
		"-n", fmt.Sprintf("%d", a.id),
		name,
	).Run()
}

func (a *app) getActionWithMenu(actions map[string]string) (string, error) {
	actionValues := make([]string, len(actions))
	reverse := make(map[string]string, len(actions))
	i := 0
	for name, action := range actions {
		actionValues[i] = action
		reverse[action] = name
		i++
	}

	cmd := exec.Command(
		"rofi",
		"-dmenu",
		"-p", "Select Action",
		"-i",
		"--only-match",
	)

	cmd.Stdin = strings.NewReader(strings.Join(actionValues, "\n"))

	data, err := cmd.Output()
	if err != nil {
		return "", err
	}

	result := strings.TrimRight(string(data), "\n")

	action, ok := reverse[result]
	if !ok {
		return "", fmt.Errorf("action not found for %q", string(data))
	}

	return action, nil
}

func (a *app) run(ctx context.Context) error {
	if err := a.flags.Parse(os.Args[1:]); err != nil {
		return err
	}

	if a.id < 0 {
		return fmt.Errorf("id is required")
	}

	notifications, err := a.getNotifications(ctx)
	if err != nil {
		return err
	}

	notification, err := notifications.GetByID(a.id)
	if err != nil {
		return err
	}

	actions := notification.Actions.Data

	switch len(actions) {
	case 0:
		return a.dismiss()
	case 1:
		for name := range actions {
			return a.invoke(name)
		}
	default:
		action, err := a.getActionWithMenu(actions)
		if err != nil {
			return err
		}

		return a.invoke(action)
	}

	return nil
}
