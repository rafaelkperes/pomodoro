package slack

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/gorilla/schema"
)

type Command struct {
	Token       string `schema:"token"`
	Command     string `schema:"command"`
	Text        string `schema:"text"`
	ResponseURL string `schema:"response_url"`
	TriggerID   string `schema:"trigger_id"`
	UserID      string `schema:"user_id"`
	UserName    string `schema:"user_name"`
}

// ParseCommand parses a Slack command request body into a struct.
func ParseCommand(r io.Reader) (Command, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return Command{}, err
	}
	// Command is always encoded as "application/x-www-form-urlencoded"
	v, err := url.ParseQuery(string(b))
	if err != nil {
		return Command{}, fmt.Errorf("could not parse as url-encoded value: %w", err)
	}
	var cmd Command
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	if err := dec.Decode(&cmd, v); err != nil {
		return Command{}, fmt.Errorf("could not parse as Slack command: %w", err)
	}
	return cmd, nil
}
