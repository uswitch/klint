package alerts

import (
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
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

	messageParameters := slack.NewPostMessageParameters()
	messageParameters.AsUser = true

	var err error = nil

	log.Debugf("sending alert \"%s\" to '%s'", message, val)

	if _, _, err = s.client.PostMessage(val, message, messageParameters); err != nil {
		log.Errorf("Failed to send message \"%s\" to '%s': %s", message, val, err)
	}

	return err
}
