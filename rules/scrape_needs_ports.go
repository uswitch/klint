package rules

import (
	log "github.com/Sirupsen/logrus"

	extv1b1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/uswitch/klint/engine"
)

func validScrapeAndPorts(d *extv1b1.Deployment) bool {
	_, hasScrapeAnnotation := d.Spec.Template.Annotations["prometheus.io.scrape"]

	hasPorts := false
	for _, c := range d.Spec.Template.Spec.Containers {
		if len(c.Ports) > 0 {
			hasPorts = true
			break
		}
	}

	return !hasScrapeAnnotation || hasPorts
}

var ScrapeNeedsPortsRule = engine.NewRule(
	func(old runtime.Object, new runtime.Object, ctx *engine.RuleHandlerContext) {
		deployment := new.(*extv1b1.Deployment)
		logger := log.WithFields(log.Fields{"name": deployment.Name, "namespace": deployment.Namespace, "rule": "ScrapeNeedsPortsRule"})

		if old == nil || validScrapeAndPorts(old.(*extv1b1.Deployment)) != validScrapeAndPorts(deployment) {
			podName := podNameForDeployment(deployment)

			if validScrapeAndPorts(deployment) { // everything is good
				if old != nil {
					ctx.Alertf(new, "Thanks for sorting the ports for scraping on %s.%s", deployment.ObjectMeta.Namespace, podName)
				}
			} else { // stuff has gone bad
				ctx.Alertf(new, "%s.%s wants to be scraped so it needs to expose some ports", deployment.ObjectMeta.Namespace, podName)
			}
		} else {
			logger.Debugf("ScrapeNeedsPortsRule: %s.%s hadn't changed", deployment.ObjectMeta.Namespace, podNameForDeployment(deployment))
		}
	},
	engine.WantDeployments,
)
