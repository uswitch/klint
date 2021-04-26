package rules

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/uswitch/klint/engine"

	batchv1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

var RequireCronJobHistoryLimits = engine.NewRule(
	func(old runtime.Object, new runtime.Object, ctx *engine.RuleHandlerContext) {
		job := new.(*batchv1.CronJob)
		logger := log.WithFields(log.Fields{"rule": "RequireCronJobHistoryLimits", "namespace": job.GetNamespace(), "name": job.GetName()})

		logger.Debugf("checking for history limit requirement")

		messages := make([]string, 0)
		if job.Spec.SuccessfulJobsHistoryLimit == nil {
			message := fmt.Sprintf("CronJob `%s/%s` doesn't specify `.spec.successfulJobsHistoryLimit`. Must be 10 or under.", job.GetNamespace(), job.GetName())
			messages = append(messages, message)
		} else {
			if *job.Spec.SuccessfulJobsHistoryLimit > 10 {
				message := fmt.Sprintf("CronJob `%s/%s` `.spec.succcessfulJobsHistoryLimit` is too high: `%d`. Must be 10 or under.", job.GetNamespace(), job.GetName(), *job.Spec.SuccessfulJobsHistoryLimit)
				messages = append(messages, message)
			}
		}

		if job.Spec.FailedJobsHistoryLimit == nil {
			message := fmt.Sprintf("CronJob `%s/%s` doesn't specify `.spec.failedJobsHistoryLimit`. Must be 10 or under.", job.GetNamespace(), job.GetName())
			messages = append(messages, message)
		} else {
			if *job.Spec.FailedJobsHistoryLimit > 10 {
				message := fmt.Sprintf("CronJob `%s/%s` `.spec.failedJobsHistoryLimit` is too high: `%d`. Must be 10 or under.", job.GetNamespace(), job.GetName(), *job.Spec.FailedJobsHistoryLimit)
				messages = append(messages, message)
			}
		}

		for _, msg := range messages {
			ctx.Alert(job, msg)
		}
	},
	engine.WantCronJobs,
)
