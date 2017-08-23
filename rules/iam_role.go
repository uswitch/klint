package rules

import (
	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/uswitch/klint/engine"

	extv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

const AnnotationName = "iam.amazonaws.com/role"

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

var ValidIAMRoleRule = engine.NewRule(
	func(old runtime.Object, new runtime.Object, ctx *engine.RuleHandlerContext) {
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
				ctx.Alertf(new, "IAM role `%s` specified for pods in deployment `%s.%s` but doesn't exist", deployment.GetNamespace(), deployment.GetName(), roleName)
				return
			}

			logger.Errorf("error finding role: %s", err.Error())
			return
		}

		logger.Debugf("iam configured correctly, huzzah!")
	},
	engine.WantDeployments,
)
