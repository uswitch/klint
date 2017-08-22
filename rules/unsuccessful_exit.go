package rules

import (
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
					ctx.Alertf(new, "Pod `%s.%s` (container: `%s`) has failed with exit code: `%d`", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, c.Name, c.State.Terminated.ExitCode)
				}
			}
		}
	},
	engine.WantPods,
)

// can we not show exit code 143 and co if the pod is terminating, it is noisy
