package alerts

import (
	log "github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
)

type SlackOutput struct {
	client   *slack.Client
}

func NewSlackOutput(token string) *SlackOutput {
	return &SlackOutput{
		client:   slack.New(token),
	}
}

func (s *SlackOutput) Key() string { return "slack" }

func (s *SlackOutput) Send(val string, message string) error {
	log.Debugf("SLACK: #%s %s", val, message)

	messageParameters := slack.NewPostMessageParameters()
	messageParameters.AsUser = true

	var err error = nil

	if _, _, err = s.client.PostMessage(val, message, messageParameters); err != nil {
		log.Warnf("Failed to send message to '%s': %s", val, err)
	}

	return err
}
