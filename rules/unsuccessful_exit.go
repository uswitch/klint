package rules

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/uswitch/klint/engine"
)

var UnsuccessfulExitRule = engine.NewRule(
	func(old runtime.Object, newObj runtime.Object, ctx *engine.RuleHandlerContext) {
		pod := newObj.(*v1.Pod)

		logger := log.WithFields(log.Fields{"name": pod.Name, "namespace": pod.Namespace, "rule": "UnsuccessfulExitRule"})

		for _, c := range pod.Status.ContainerStatuses {
			logger = logger.WithFields(log.Fields{"container.name": c.Name, "container.id": c.ContainerID})

			if c.State.Terminated != nil {
				switch exitCode := c.State.Terminated.ExitCode; exitCode {
				case 0: // Everything was OK
					break
				case 143: // JVM SIGTERM
					break
				case 137: // Process got SIGKILLd
					ctx.Alertf(newObj, "Pod `%s.%s` (container: `%s`) was killed by a SIGKILL. Please make sure you gracefully shut down in time or extend `terminationGracePeriodSeconds` on your pod.", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, c.Name)
				default:
					tailLines := int64(20)
					opts := &v1.PodLogOptions{
						Container: c.Name,
						Follow:    false,
						Previous:  false,
						TailLines: &tailLines,
					}
					message := fmt.Sprintf("Pod `%s.%s` (container: `%s`) has failed with exit code: `%d`", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, c.Name, c.State.Terminated.ExitCode)

					result := ctx.Client().CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, opts).Do()
					if result.Error() != nil {
						logger.Errorf("error retrieving pod logs: %s", result.Error())
						ctx.Alert(newObj, message)
						return
					}

					bytes, err := result.Raw()
					if err != nil {
						logger.Errorf("error retrieving pod logs: %s", err.Error())
						ctx.Alert(newObj, message)
						return
					}

					logger.Debugf("log: \"%s\"", string(bytes))
					ctx.Alertf(newObj, "%s\n\n```%s```", message, string(bytes))
				}
			}
		}
	},
	engine.WantPods,
)

// can we not show exit code 143 and co if the pod is terminating, it is noisy
