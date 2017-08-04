package alerts

import (
	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type SNSOutput struct {
	client *sns.SNS
}

func NewSNSOutput(region string) *SNSOutput {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	return &SNSOutput{
		client: sns.New(sess),
	}
}

func (s *SNSOutput) Key() string { return "sns" }

func (s *SNSOutput) Send(val string, message string) error {
	log.Debugf("SNS: #%s %s", val, message)

	params := &sns.PublishInput{
		Subject:  aws.String(message),
		Message:  aws.String(message),
		TopicArn: aws.String(val),
	}
	_, err := s.client.Publish(params)

	return err
}
