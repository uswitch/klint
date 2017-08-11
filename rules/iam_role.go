package rules

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/uswitch/klint/alerts"

	extv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

const AnnotationName = "iam.amazonaws.com/role"

func alertNoRole(deployment *extv1.Deployment, out chan *alerts.Alert) {
	roleName := role(deployment)
	message := fmt.Sprintf("IAM role %s specified for pods but doesn't exist", roleName)
	out <- &alerts.Alert{
		deployment,
		message,
	}
}

func role(deployment *extv1.Deployment) string {
	return deployment.Spec.Template.GetAnnotations()[AnnotationName]
}

func fields(deployment *extv1.Deployment) log.Fields {
	return log.Fields{
		"namespace": deployment.GetNamespace(),
		"name":      deployment.GetName(),
		"role":      role(deployment),
	}
}

var ValidIAMRoleRule = NewRule(
	func(old runtime.Object, new runtime.Object, out chan *alerts.Alert) {
		deployment := new.(*extv1.Deployment)
		logger := log.WithFields(fields(deployment))

		logger.Debugf("checking deployment for iam infringement")

		roleName := role(deployment)
		if roleName == "" {
			return
		}

		session := session.New()
		svc := iam.New(session)

		_, err := svc.GetRole(&iam.GetRoleInput{RoleName: aws.String(roleName)})
		if err != nil {
			e, _ := err.(awserr.Error)
			if e.Code() == iam.ErrCodeNoSuchEntityException {
				alertNoRole(deployment, out)
				return
			}

			logger.Errorf("error finding role: %s", err.Error())
			return
		}

		logger.Debugf("iam configured correctly, huzzah!")
	},
	WantDeployments,
)
