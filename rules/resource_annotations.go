package rules

import (
	"reflect"
	"strings"

	log "github.com/Sirupsen/logrus"

	"k8s.io/api/core/v1"
	extv1b1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/uswitch/klint/engine"
)

func hasKeys(m v1.ResourceList, keys ...string) bool {
	for _, key := range keys {
		if _, ok := m[v1.ResourceName(key)]; !ok {
			return false
		}
	}

	return true
}

func podNameForDeployment(deployment *extv1b1.Deployment) string {
	podName := deployment.Spec.Template.ObjectMeta.Name
	if podName == "" {
		podName = deployment.ObjectMeta.Name
	}
	return podName
}

func containersInViolation(deployment *extv1b1.Deployment) []string {
	containersMissingResources := []string{}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		if !(hasKeys(container.Resources.Requests, "cpu", "memory") && hasKeys(container.Resources.Limits, "memory")) {
			containersMissingResources = append(containersMissingResources, container.Name)
		}
	}

	return containersMissingResources
}

var ResourceAnnotationRule = engine.NewRule(
	func(old runtime.Object, new runtime.Object, ctx *engine.RuleHandlerContext) {
		deployment := new.(*extv1b1.Deployment)
		logger := log.WithFields(log.Fields{"rule": "ResourceAnnotationRule", "name": deployment.Name, "namespace": deployment.Namespace})

		newInViolation := containersInViolation(deployment)

		if old == nil || !reflect.DeepEqual(containersInViolation(old.(*extv1b1.Deployment)), newInViolation) {
			if len(newInViolation) == 0 { // it wasn't zero before so they've fixed their issues
				if old != nil {
					ctx.Alertf(new, "Thanks for sorting your resource requests and limits on %s.%s!", deployment.ObjectMeta.Namespace, podNameForDeployment(deployment))
				}
			} else { // it's now more or less broken than it was before, but not fixed
				ctx.Alertf(new, "Please add resource requests and limits to the containers (%s) part of %s.%s", strings.Join(newInViolation, ", "), deployment.ObjectMeta.Namespace, podNameForDeployment(deployment))
			}
		} else {
			logger.Debugf("ResourceAnnotationRule: %s.%s hadn't changed", deployment.ObjectMeta.Namespace, podNameForDeployment(deployment))
		}
	},
	engine.WantDeployments,
)
