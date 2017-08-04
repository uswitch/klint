package rules

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/uswitch/klint/alerts"
)

var UnsuccessfulExitRule = NewRule(
	func (old runtime.Object, new runtime.Object, out chan *alerts.Alert) {
		pod := new.(*v1.Pod)

		for _, c := range pod.Status.ContainerStatuses {
			if c.State.Terminated != nil && c.State.Terminated.ExitCode > 0 {
				log.Debugf("Unhappy pod %s", pod.ObjectMeta.Name)

				out <- &alerts.Alert{
					new,
					fmt.Sprintf(
						"Pod `%s.%s` (container: `%s`) has failed with exit code: `%d`",
						pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, c.Name, c.State.Terminated.ExitCode,
					),
				}
			}
		}
	},
	WantPods,
)

// can we not show exit code 143 and co if the pod is terminating, it is noisy
