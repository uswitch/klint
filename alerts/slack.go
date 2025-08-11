package alerts

import (
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
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

	var err error = nil

	log.Debugf("sending alert \"%s\" to '%s'", message, val)

	if _, _, err = s.client.PostMessage(val, slack.MsgOptionText(message, false)); err != nil {
		log.Errorf("Failed to send message \"%s\" to '%s': %s", message, val, err)
	}

	return err
}
