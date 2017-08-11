package rules

import (
	"fmt"

	//log "github.com/Sirupsen/logrus"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/uswitch/klint/alerts"
)

var UnsuccessfulExitRule = NewRule(
	func(old runtime.Object, new runtime.Object, out chan *alerts.Alert) {
		pod := new.(*v1.Pod)

		for _, c := range pod.Status.ContainerStatuses {
			if c.State.Terminated != nil {
				switch exitCode := c.State.Terminated.ExitCode; exitCode {
				case 0:    // Everything was OK
					break
				case 143:  // JVM SIGTERM
					break
				case 137:  // Process got SIGKILLd
					out <- &alerts.Alert{
						new,
						fmt.Sprintf(
							"Pod `%s.%s` (container: `%s`) was killed by a SIGKILL. Please make sure you are shuting down in time, or extend `terminationGracePeriodSeconds` on your pod.",
							pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, c.Name,
						),
					}
				default:
					out <- &alerts.Alert{
						new,
						fmt.Sprintf(
							"Pod `%s.%s` (container: `%s`) has failed with exit code: `%d`",
							pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, c.Name, c.State.Terminated.ExitCode,
						),
					}
				}

			}
		}
	},
	WantPods,
)

// can we not show exit code 143 and co if the pod is terminating, it is noisy
