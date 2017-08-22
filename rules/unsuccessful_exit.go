package rules

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/uswitch/klint/engine"
)

var UnsuccessfulExitRule = engine.NewRule(
	func(old runtime.Object, new runtime.Object, ctx *engine.RuleHandlerContext) {
		pod := new.(*v1.Pod)

		for _, c := range pod.Status.ContainerStatuses {
			if c.State.Terminated != nil {
				switch exitCode := c.State.Terminated.ExitCode; exitCode {
				case 0: // Everything was OK
					break
				case 143: // JVM SIGTERM
					break
				case 137: // Process got SIGKILLd
					ctx.Alertf(new, "Pod `%s.%s` (container: `%s`) was killed by a SIGKILL. Please make sure you gracefully shut down in time or extend `terminationGracePeriodSeconds` on your pod.", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, c.Name)
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
						log.Errorf("error retrieving pod logs: %s", err.Error())
						ctx.Alert(new, message)
						return
					}

					defer rs.Close()
					scanner := bufio.NewScanner(rs)
					for scanner.Scan() {
						log.Debugf("log: %s", scanner.Text())
					}

					ctx.Alert(new, message)
				}
			}
		}
	},
	engine.WantPods,
)

// can we not show exit code 143 and co if the pod is terminating, it is noisy
