package alerts

import (
	"github.com/nlopes/slack"
	log "github.com/Sirupsen/logrus"
)


type SlackOutput struct {
	client *slack.Client
}

func NewSlackOutput(token string) *SlackOutput {
	return &SlackOutput{
		client: slack.New(token),
	}
}

func (s *SlackOutput) Key() string { return "slack" }

func (s *SlackOutput) Send(val string, message string) error {
	log.Debugf("SLACK: #%s %s", val, message)
	_, _, err := s.client.PostMessage(val, message, slack.PostMessageParameters{})

	return err
}
