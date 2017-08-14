package rules

import (
	"fmt"
	log "github.com/Sirupsen/logrus"

	"github.com/uswitch/klint/alerts"

	batchv2 "k8s.io/api/batch/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func cronjobFields(job *batchv2.CronJob) log.Fields {
	return log.Fields{
		"namespace": job.GetNamespace(),
		"name":      job.GetName(),
	}
}

var RequireCronJobHistoryLimits = NewRule(
	func(old runtime.Object, new runtime.Object, out chan *alerts.Alert) {
		job := new.(*batchv2.CronJob)
		logger := log.WithFields(cronjobFields(job))

		logger.Debugf("checking for history limit requirement")

		messages := make([]string, 0)
		if job.Spec.SuccessfulJobsHistoryLimit == nil {
			message := fmt.Sprintf("CronJob `%s/%s` doesn't specify successfulJobsHistoryLimit. Must be 10 or under.", job.GetNamespace(), job.GetName())
			messages = append(messages, message)
		} else {
			if *job.Spec.SuccessfulJobsHistoryLimit > 10 {
				message := fmt.Sprintf("CronJob `%s/%s` succcessfulJobsHistoryLimit is too high: `%d`. Must be 10 or under.", job.GetNamespace(), job.GetName(), *job.Spec.SuccessfulJobsHistoryLimit)
				messages = append(messages, message)
			}
		}

		if job.Spec.FailedJobsHistoryLimit == nil {
			message := fmt.Sprintf("CronJob `%s/%s` doesn't specify failedJobsHistoryLimit. Must be 10 or under.", job.GetNamespace(), job.GetName())
			messages = append(messages, message)
		} else {
			if *job.Spec.FailedJobsHistoryLimit > 10 {
				message := fmt.Sprintf("CronJob `%s/%s` failedJobsHistoryLimit is too high: `%d`. Must be 10 or under.", job.GetNamespace(), job.GetName(), *job.Spec.FailedJobsHistoryLimit)
				messages = append(messages, message)
			}
		}

		for _, msg := range messages {
			out <- &alerts.Alert{job, msg}
		}
	},
	WantCronJobs,
)
