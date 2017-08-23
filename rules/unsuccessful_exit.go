package rules

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/uswitch/klint/engine"
)

var UnsuccessfulExitRule = engine.NewRule(
	func(old runtime.Object, newObj runtime.Object, ctx *engine.RuleHandlerContext) {
		pod := newObj.(*v1.Pod)

		logger := log.WithFields(log.Fields{"name": pod.Name, "namespace": pod.Namespace, "rule": "UnsuccessfulExitRule"})

		for _, c := range pod.Status.ContainerStatuses {
			if c.State.Terminated != nil {
				switch exitCode := c.State.Terminated.ExitCode; exitCode {
				case 0: // Everything was OK
					break
				case 143: // JVM SIGTERM
					break
				case 137: // Process got SIGKILLd
					ctx.Alertf(newObj, "Pod `%s.%s` (container: `%s`) was killed by a SIGKILL. Please make sure you gracefully shut down in time or extend `terminationGracePeriodSeconds` on your pod.", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, c.Name)
				default:
					since := int64(30)
					opts := &v1.PodLogOptions{
						Container:    c.Name,
						Follow:       false,
						SinceSeconds: &since,
					}
					message := fmt.Sprintf("Pod `%s.%s` (container: `%s`) has failed with exit code: `%d`", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, c.Name, c.State.Terminated.ExitCode)

					rs, err := ctx.Client().Core().Pods(pod.Namespace).GetLogs(pod.Name, opts).Stream()
					if err != nil {
						logger.Errorf("error retrieving pod logs: %s", err.Error())
						ctx.Alert(newObj, message)
						return
					}

					defer rs.Close()
					buf := new(bytes.Buffer)
					_, err = buf.ReadFrom(rs)
					if err != nil {
						logger.Errorf("error reading log stream: %s", err.Error())
						// we weren't able to extract logs but we should still report errors
						ctx.Alert(newObj, message)
						return
					}

					logger.Debugf("log: \"%s\"", buf.String())
					ctx.Alertf(newObj, "%s\n```%s```", message, buf.String())
				}
			}
		}
	},
	engine.WantPods,
)

// can we not show exit code 143 and co if the pod is terminating, it is noisy
