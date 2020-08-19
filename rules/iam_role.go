package rules

import (
	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/uswitch/klint/engine"

	appsv1 "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

const AnnotationName = "iam.amazonaws.com/role"

var ValidIAMRoleRule = engine.NewRule(
	func(old runtime.Object, new runtime.Object, ctx *engine.RuleHandlerContext) {
		deployment := new.(*appsv1.Deployment)
		roleName := deployment.Spec.Template.GetAnnotations()[AnnotationName]
		logger := log.WithFields(log.Fields{"rule": "ValidIAMRoleRule", "namespace": deployment.GetNamespace(), "name": deployment.GetName(), "role": roleName})

		logger.Debugf("checking deployment for iam infringement")
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
